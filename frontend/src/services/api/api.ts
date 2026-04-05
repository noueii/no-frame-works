import { api } from "./client";
export const addTagTypes = ["auth", "users", "posts"] as const;
const injectedRtkApi = api
  .enhanceEndpoints({
    addTagTypes,
  })
  .injectEndpoints({
    endpoints: (build) => ({
      postAuthLogin: build.mutation<
        PostAuthLoginApiResponse,
        PostAuthLoginApiArg
      >({
        query: (queryArg) => ({
          url: `/auth/login`,
          method: "POST",
          body: queryArg.loginRequest,
        }),
        invalidatesTags: ["auth"],
      }),
      postAuthRegister: build.mutation<
        PostAuthRegisterApiResponse,
        PostAuthRegisterApiArg
      >({
        query: (queryArg) => ({
          url: `/auth/register`,
          method: "POST",
          body: queryArg.registerRequest,
        }),
        invalidatesTags: ["auth"],
      }),
      postAuthLogout: build.mutation<
        PostAuthLogoutApiResponse,
        PostAuthLogoutApiArg
      >({
        query: () => ({ url: `/auth/logout`, method: "POST" }),
        invalidatesTags: ["auth"],
      }),
      postUsers: build.mutation<PostUsersApiResponse, PostUsersApiArg>({
        query: (queryArg) => ({
          url: `/users`,
          method: "POST",
          body: queryArg.createUserRequest,
        }),
        invalidatesTags: ["users"],
      }),
      getUsersById: build.query<GetUsersByIdApiResponse, GetUsersByIdApiArg>({
        query: (queryArg) => ({ url: `/users/${queryArg.id}` }),
        providesTags: ["users"],
      }),
      postPosts: build.mutation<PostPostsApiResponse, PostPostsApiArg>({
        query: (queryArg) => ({
          url: `/posts`,
          method: "POST",
          body: queryArg.createPostRequest,
        }),
        invalidatesTags: ["posts"],
      }),
      getPostsById: build.query<GetPostsByIdApiResponse, GetPostsByIdApiArg>({
        query: (queryArg) => ({ url: `/posts/${queryArg.id}` }),
        providesTags: ["posts"],
      }),
      getPostsByAuthorByAuthorId: build.query<
        GetPostsByAuthorByAuthorIdApiResponse,
        GetPostsByAuthorByAuthorIdApiArg
      >({
        query: (queryArg) => ({ url: `/posts/by-author/${queryArg.authorId}` }),
        providesTags: ["posts"],
      }),
    }),
    overrideExisting: false,
  });
export { injectedRtkApi as api };
export type PostAuthLoginApiResponse =
  /** status 200 Login successful */ SessionResponse;
export type PostAuthLoginApiArg = {
  loginRequest: LoginRequest;
};
export type PostAuthRegisterApiResponse =
  /** status 200 Registration successful */ SessionResponse;
export type PostAuthRegisterApiArg = {
  registerRequest: RegisterRequest;
};
export type PostAuthLogoutApiResponse = unknown;
export type PostAuthLogoutApiArg = void;
export type PostUsersApiResponse = /** status 201 User created */ User;
export type PostUsersApiArg = {
  createUserRequest: CreateUserRequest;
};
export type GetUsersByIdApiResponse = /** status 200 User found */ User;
export type GetUsersByIdApiArg = {
  id: string;
};
export type PostPostsApiResponse = /** status 201 Post created */ Post;
export type PostPostsApiArg = {
  createPostRequest: CreatePostRequest;
};
export type GetPostsByIdApiResponse = /** status 200 Post found */ Post;
export type GetPostsByIdApiArg = {
  id: string;
};
export type GetPostsByAuthorByAuthorIdApiResponse =
  /** status 200 Posts found */ Post[];
export type GetPostsByAuthorByAuthorIdApiArg = {
  authorId: string;
};
export type SessionResponse = {
  sessionToken: string;
};
export type LoginRequest = {
  email: string;
  password: string;
};
export type RegisterRequest = {
  email: string;
  password: string;
};
export type User = {
  id: string;
  name: string;
  email: string;
};
export type CreateUserRequest = {
  name: string;
  email: string;
};
export type Post = {
  id: string;
  title: string;
  content: string;
  authorId: string;
  authorName: string;
};
export type CreatePostRequest = {
  title: string;
  content: string;
};
export const {
  usePostAuthLoginMutation,
  usePostAuthRegisterMutation,
  usePostAuthLogoutMutation,
  usePostUsersMutation,
  useGetUsersByIdQuery,
  useLazyGetUsersByIdQuery,
  usePostPostsMutation,
  useGetPostsByIdQuery,
  useLazyGetPostsByIdQuery,
  useGetPostsByAuthorByAuthorIdQuery,
  useLazyGetPostsByAuthorByAuthorIdQuery,
} = injectedRtkApi;
