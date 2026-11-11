/* eslint-disable */
import * as types from './graphql';
import type { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';

/**
 * Map of all GraphQL operations in the project.
 *
 * This map has several performance disadvantages:
 * 1. It is not tree-shakeable, so it will include all operations in the project.
 * 2. It is not minifiable, so the string of a GraphQL query will be multiple times inside the bundle.
 * 3. It does not support dead code elimination, so it will add unused operations.
 *
 * Therefore it is highly recommended to use the babel or swc plugin for production.
 * Learn more about it here: https://the-guild.dev/graphql/codegen/plugins/presets/preset-client#reducing-bundle-size
 */
type Documents = {
    "\n  query Health {\n    health\n  }\n": typeof types.HealthDocument,
    "\n  mutation Logout {\n    logout\n  }\n": typeof types.LogoutDocument,
    "\n  mutation UpdateDisplayName($input: UpdateDisplayNameInput!) {\n    updateDisplayName(input: $input) {\n      viewer {\n        id\n        displayName\n      }\n      errors {\n        field\n        code\n      }\n    }\n  }\n": typeof types.UpdateDisplayNameDocument,
    "\n  mutation ChangePassword($input: ChangePasswordInput!) {\n    changePassword(input: $input) {\n      viewer {\n        id\n      }\n      errors {\n        field\n        code\n      }\n    }\n  }\n": typeof types.ChangePasswordDocument,
    "\n  mutation IssueInvitation {\n    issueInvitation {\n      code\n      expiresAt\n    }\n  }\n": typeof types.IssueInvitationDocument,
    "\n  mutation Login($input: LoginInput!) {\n    login(input: $input) {\n      viewer {\n        id\n      }\n      errors {\n        field\n        code\n      }\n      lockedUntil\n    }\n  }\n": typeof types.LoginDocument,
    "\n  mutation Register($input: RegisterInput!) {\n    register(input: $input) {\n      viewer {\n        id\n      }\n      errors {\n        field\n        code\n      }\n    }\n  }\n": typeof types.RegisterDocument,
    "\n  mutation ResetPassword($input: ResetPasswordInput!) {\n    resetPassword(input: $input) {\n      ok\n      errors {\n        field\n        code\n      }\n      lockedUntil\n    }\n  }\n": typeof types.ResetPasswordDocument,
    "\n  query Viewer {\n    viewer {\n      id\n      loginId\n      displayName\n      isAdmin\n    }\n  }\n": typeof types.ViewerDocument,
};
const documents: Documents = {
    "\n  query Health {\n    health\n  }\n": types.HealthDocument,
    "\n  mutation Logout {\n    logout\n  }\n": types.LogoutDocument,
    "\n  mutation UpdateDisplayName($input: UpdateDisplayNameInput!) {\n    updateDisplayName(input: $input) {\n      viewer {\n        id\n        displayName\n      }\n      errors {\n        field\n        code\n      }\n    }\n  }\n": types.UpdateDisplayNameDocument,
    "\n  mutation ChangePassword($input: ChangePasswordInput!) {\n    changePassword(input: $input) {\n      viewer {\n        id\n      }\n      errors {\n        field\n        code\n      }\n    }\n  }\n": types.ChangePasswordDocument,
    "\n  mutation IssueInvitation {\n    issueInvitation {\n      code\n      expiresAt\n    }\n  }\n": types.IssueInvitationDocument,
    "\n  mutation Login($input: LoginInput!) {\n    login(input: $input) {\n      viewer {\n        id\n      }\n      errors {\n        field\n        code\n      }\n      lockedUntil\n    }\n  }\n": types.LoginDocument,
    "\n  mutation Register($input: RegisterInput!) {\n    register(input: $input) {\n      viewer {\n        id\n      }\n      errors {\n        field\n        code\n      }\n    }\n  }\n": types.RegisterDocument,
    "\n  mutation ResetPassword($input: ResetPasswordInput!) {\n    resetPassword(input: $input) {\n      ok\n      errors {\n        field\n        code\n      }\n      lockedUntil\n    }\n  }\n": types.ResetPasswordDocument,
    "\n  query Viewer {\n    viewer {\n      id\n      loginId\n      displayName\n      isAdmin\n    }\n  }\n": types.ViewerDocument,
};

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 *
 *
 * @example
 * ```ts
 * const query = graphql(`query GetUser($id: ID!) { user(id: $id) { name } }`);
 * ```
 *
 * The query argument is unknown!
 * Please regenerate the types.
 */
export function graphql(source: string): unknown;

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Health {\n    health\n  }\n"): (typeof documents)["\n  query Health {\n    health\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation Logout {\n    logout\n  }\n"): (typeof documents)["\n  mutation Logout {\n    logout\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation UpdateDisplayName($input: UpdateDisplayNameInput!) {\n    updateDisplayName(input: $input) {\n      viewer {\n        id\n        displayName\n      }\n      errors {\n        field\n        code\n      }\n    }\n  }\n"): (typeof documents)["\n  mutation UpdateDisplayName($input: UpdateDisplayNameInput!) {\n    updateDisplayName(input: $input) {\n      viewer {\n        id\n        displayName\n      }\n      errors {\n        field\n        code\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation ChangePassword($input: ChangePasswordInput!) {\n    changePassword(input: $input) {\n      viewer {\n        id\n      }\n      errors {\n        field\n        code\n      }\n    }\n  }\n"): (typeof documents)["\n  mutation ChangePassword($input: ChangePasswordInput!) {\n    changePassword(input: $input) {\n      viewer {\n        id\n      }\n      errors {\n        field\n        code\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation IssueInvitation {\n    issueInvitation {\n      code\n      expiresAt\n    }\n  }\n"): (typeof documents)["\n  mutation IssueInvitation {\n    issueInvitation {\n      code\n      expiresAt\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation Login($input: LoginInput!) {\n    login(input: $input) {\n      viewer {\n        id\n      }\n      errors {\n        field\n        code\n      }\n      lockedUntil\n    }\n  }\n"): (typeof documents)["\n  mutation Login($input: LoginInput!) {\n    login(input: $input) {\n      viewer {\n        id\n      }\n      errors {\n        field\n        code\n      }\n      lockedUntil\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation Register($input: RegisterInput!) {\n    register(input: $input) {\n      viewer {\n        id\n      }\n      errors {\n        field\n        code\n      }\n    }\n  }\n"): (typeof documents)["\n  mutation Register($input: RegisterInput!) {\n    register(input: $input) {\n      viewer {\n        id\n      }\n      errors {\n        field\n        code\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation ResetPassword($input: ResetPasswordInput!) {\n    resetPassword(input: $input) {\n      ok\n      errors {\n        field\n        code\n      }\n      lockedUntil\n    }\n  }\n"): (typeof documents)["\n  mutation ResetPassword($input: ResetPasswordInput!) {\n    resetPassword(input: $input) {\n      ok\n      errors {\n        field\n        code\n      }\n      lockedUntil\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Viewer {\n    viewer {\n      id\n      loginId\n      displayName\n      isAdmin\n    }\n  }\n"): (typeof documents)["\n  query Viewer {\n    viewer {\n      id\n      loginId\n      displayName\n      isAdmin\n    }\n  }\n"];

export function graphql(source: string) {
  return (documents as any)[source] ?? {};
}

export type DocumentType<TDocumentNode extends DocumentNode<any, any>> = TDocumentNode extends DocumentNode<  infer TType,  any>  ? TType  : never;