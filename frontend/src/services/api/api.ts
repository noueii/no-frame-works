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
      getAuthMe: build.query<GetAuthMeApiResponse, GetAuthMeApiArg>({
        query: () => ({ url: `/auth/me` }),
        providesTags: ["auth"],
      }),
      getUsersById: build.query<GetUsersByIdApiResponse, GetUsersByIdApiArg>({
        query: (queryArg) => ({ url: `/users/${queryArg.id}` }),
        providesTags: ["users"],
      }),
      getPosts: build.query<GetPostsApiResponse, GetPostsApiArg>({
        query: (queryArg) => ({
          url: `/posts`,
          params: {
            authorId: queryArg.authorId,
          },
        }),
        providesTags: ["posts"],
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
      putPostsById: build.mutation<PutPostsByIdApiResponse, PutPostsByIdApiArg>(
        {
          query: (queryArg) => ({
            url: `/posts/${queryArg.id}`,
            method: "PUT",
            body: queryArg.updatePostRequest,
          }),
          invalidatesTags: ["posts"],
        },
      ),
      deletePostsById: build.mutation<
        DeletePostsByIdApiResponse,
        DeletePostsByIdApiArg
      >({
        query: (queryArg) => ({
          url: `/posts/${queryArg.id}`,
          method: "DELETE",
        }),
        invalidatesTags: ["posts"],
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
export type GetAuthMeApiResponse =
  /** status 200 Authenticated user */ AuthUser;
export type GetAuthMeApiArg = void;
export type GetUsersByIdApiResponse = /** status 200 User found */ User;
export type GetUsersByIdApiArg = {
  id: string;
};
export type GetPostsApiResponse = /** status 200 Posts */ Post[];
export type GetPostsApiArg = {
  /** Filter posts by author ID */
  authorId?: string;
};
export type PostPostsApiResponse = /** status 201 Post created */ Post;
export type PostPostsApiArg = {
  createPostRequest: CreatePostRequest;
};
export type GetPostsByIdApiResponse = /** status 200 Post found */ Post;
export type GetPostsByIdApiArg = {
  id: string;
};
export type PutPostsByIdApiResponse = /** status 200 Post updated */ Post;
export type PutPostsByIdApiArg = {
  id: string;
  updatePostRequest: UpdatePostRequest;
};
export type DeletePostsByIdApiResponse = unknown;
export type DeletePostsByIdApiArg = {
  id: string;
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
export type AuthUser = {
  id: string;
  email: string;
};
export type User = {
  id: string;
  email: string;
};
export type Post = {
  id: string;
  title: string;
  content: string;
  authorId: string;
};
export type CreatePostRequest = {
  title: string;
  content: string;
};
export type UpdatePostRequest = {
  title: string;
  content: string;
};
export const {
  usePostAuthLoginMutation,
  usePostAuthRegisterMutation,
  usePostAuthLogoutMutation,
  useGetAuthMeQuery,
  useLazyGetAuthMeQuery,
  useGetUsersByIdQuery,
  useLazyGetUsersByIdQuery,
  useGetPostsQuery,
  useLazyGetPostsQuery,
  usePostPostsMutation,
  useGetPostsByIdQuery,
  useLazyGetPostsByIdQuery,
  usePutPostsByIdMutation,
  useDeletePostsByIdMutation,
} = injectedRtkApi;
