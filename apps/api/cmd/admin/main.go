// Command admin は管理者の操作を行うコマンド。Taskfile の admin:* から呼ぶ
// （specs/002-auth/contracts/admin-commands.md）。
//
// 接続先は DATABASE_URL。web と api は経由せず、DB を直接操作する。
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/term"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
)

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr, os.Getenv))
}

const usage = `使い方: admin <サブコマンド> [引数]

サブコマンド:
  create-admin -login-id <ID> -display-name <表示名>   管理者を作る（パスワードは入力を求める）
  users                                                ユーザーの一覧
  reset-code <ログイン ID>                             再設定コードを発行する
  disable <ログイン ID>                                ユーザーを無効にする
  enable <ログイン ID>                                 ユーザーを有効に戻す
`

// cli は 1 回の実行の入出力。
type cli struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	svc    *auth.Service
}

func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer, getenv func(string) string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(stderr, usage)
		return 1
	}
	sub, rest := args[0], args[1:]
	commands := map[string]func(context.Context, *cli, []string) error{
		"create-admin": createAdmin,
		"users":        listUsers,
		"reset-code":   resetCode,
		"disable":      disableUser,
		"enable":       enableUser,
	}
	cmd, ok := commands[sub]
	if !ok {
		_, _ = fmt.Fprintf(stderr, "不明なサブコマンドです: %s\n\n%s", sub, usage)
		return 1
	}

	pool, err := connect(ctx, getenv("DATABASE_URL"), stderr)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	defer pool.Close()

	c := &cli{
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		svc:    auth.NewService(pool, slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))),
	}
	if err := cmd(ctx, c, rest); err != nil {
		_, _ = fmt.Fprintln(stderr, describe(err))
		return 1
	}
	return 0
}

// connect は DB に接続し、接続先（ホストとデータベース名）を表示する。
// 本番の DB につないでいることに気付けるようにする。ユーザー名とパスワードは出さない。
func connect(ctx context.Context, url string, stderr io.Writer) (*pgxpool.Pool, error) {
	if url == "" {
		return nil, errors.New("DATABASE_URL が設定されていません")
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, errors.New("DATABASE_URL の形式が不正です")
	}
	_, _ = fmt.Fprintf(stderr, "接続先: %s/%s\n", cfg.ConnConfig.Host, cfg.ConnConfig.Database)
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("DB に接続できません: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("DB に接続できません: %w", err)
	}
	return pool, nil
}

func createAdmin(ctx context.Context, c *cli, args []string) error {
	fs := flag.NewFlagSet("create-admin", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	loginID := fs.String("login-id", "", "ログイン ID（半角英数字・_・- の 3〜32 文字）")
	displayName := fs.String("display-name", "", "表示名（1〜32 文字）")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *loginID == "" || *displayName == "" {
		return errors.New("-login-id と -display-name を指定してください")
	}
	password, err := c.readNewPassword()
	if err != nil {
		return err
	}
	u, err := c.svc.CreateAdmin(ctx, *loginID, *displayName, password)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(c.stdout, "管理者を作りました: %s\n", u.LoginID)
	return nil
}

func listUsers(ctx context.Context, c *cli, _ []string) error {
	users, err := c.svc.ListUsers(ctx)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(c.stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ログイン ID\t表示名\t管理者\t状態\t作成日時")
	for _, u := range users {
		admin := ""
		if u.IsAdmin {
			admin = "管理者"
		}
		state := "有効"
		if u.DisabledAt != nil {
			state = "無効"
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", u.LoginID, u.DisplayName, admin, state, formatTime(u.CreatedAt))
	}
	return w.Flush()
}

func resetCode(ctx context.Context, c *cli, args []string) error {
	loginID, err := oneLoginID(args)
	if err != nil {
		return err
	}
	code, expiresAt, err := c.svc.IssueResetCode(ctx, loginID)
	if errors.Is(err, auth.ErrUserDisabled) {
		return fmt.Errorf("無効なユーザーには再設定コードを発行できません: %s（先に admin:enable を実行してください）", loginID)
	}
	if err != nil {
		return withLoginID(err, loginID)
	}
	_, _ = fmt.Fprintf(c.stdout, "再設定コード: %s\n有効期限: %s\n", code, formatTime(expiresAt))
	return nil
}

func disableUser(ctx context.Context, c *cli, args []string) error {
	loginID, err := oneLoginID(args)
	if err != nil {
		return err
	}
	changed, sessions, err := c.svc.DisableUser(ctx, loginID)
	if err != nil {
		return withLoginID(err, loginID)
	}
	if !changed {
		_, _ = fmt.Fprintf(c.stdout, "既に無効です: %s\n", loginID)
		return nil
	}
	_, _ = fmt.Fprintf(c.stdout, "無効にしました: %s（切ったセッション: %d）\n", loginID, sessions)
	return nil
}

func enableUser(ctx context.Context, c *cli, args []string) error {
	loginID, err := oneLoginID(args)
	if err != nil {
		return err
	}
	changed, err := c.svc.EnableUser(ctx, loginID)
	if err != nil {
		return withLoginID(err, loginID)
	}
	if !changed {
		_, _ = fmt.Fprintf(c.stdout, "既に有効です: %s\n", loginID)
		return nil
	}
	_, _ = fmt.Fprintf(c.stdout, "有効にしました: %s\n", loginID)
	return nil
}

func oneLoginID(args []string) (string, error) {
	if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
		return "", errors.New("ログイン ID を 1 つ指定してください")
	}
	return strings.TrimSpace(args[0]), nil
}

func withLoginID(err error, loginID string) error {
	if errors.Is(err, auth.ErrUserNotFound) {
		return fmt.Errorf("ユーザーが見つかりません: %s", loginID)
	}
	return err
}

// readNewPassword はパスワードを 2 回入力させる。端末なら画面に出さずに読み、
// 端末でなければ（テストなど）標準入力から 2 行で読む。引数や環境変数では受け取らない（FR-041）。
func (c *cli) readNewPassword() (string, error) {
	read := c.lineReader()
	first, err := read("パスワード: ")
	if err != nil {
		return "", err
	}
	second, err := read("パスワード（確認）: ")
	if err != nil {
		return "", err
	}
	if first != second {
		return "", errors.New("パスワードが一致しません")
	}
	return first, nil
}

func (c *cli) lineReader() func(prompt string) (string, error) {
	if f, ok := c.stdin.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		return func(prompt string) (string, error) {
			_, _ = fmt.Fprint(c.stderr, prompt)
			b, err := term.ReadPassword(int(f.Fd()))
			_, _ = fmt.Fprintln(c.stderr)
			if err != nil {
				return "", fmt.Errorf("パスワードを読めません: %w", err)
			}
			return string(b), nil
		}
	}
	sc := bufio.NewScanner(c.stdin)
	return func(string) (string, error) {
		if !sc.Scan() {
			return "", errors.New("パスワードを読めません（標準入力が空です）")
		}
		return strings.TrimRight(sc.Text(), "\r"), nil
	}
}

func formatTime(t time.Time) string {
	return t.Local().Format("2006-01-02 15:04")
}

// describe はエラーを利用者向けの文言にする。
func describe(err error) string {
	var ie auth.InputErrors
	if !errors.As(err, &ie) {
		return err.Error()
	}
	msgs := make([]string, len(ie))
	for i, e := range ie {
		msgs[i] = inputMessage(e)
	}
	return strings.Join(msgs, "\n")
}

func inputMessage(e auth.InputError) string {
	switch e.Field {
	case auth.FieldLoginID:
		if e.Code == auth.CodeLoginIDTaken {
			return "ログイン ID が既に使われています"
		}
		return fmt.Sprintf("ログイン ID は半角英数字・_・- の %d〜%d 文字にしてください", auth.LoginIDMinLen, auth.LoginIDMaxLen)
	case auth.FieldPassword:
		return fmt.Sprintf("パスワードは %d〜%d 文字にしてください", auth.PasswordMinLen, auth.PasswordMaxLen)
	case auth.FieldDisplayName:
		return fmt.Sprintf("表示名は改行などを含まない 1〜%d 文字にしてください", auth.DisplayNameMaxLen)
	}
	return fmt.Sprintf("入力が不正です（%s: %s）", e.Field, e.Code)
}
