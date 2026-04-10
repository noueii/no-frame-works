import { useState } from 'react'
import { useAuth } from '../contexts/AuthContext'
import {
  usePostPostsMutation,
  useGetPostsQuery,
  usePutPostsByIdMutation,
  useDeletePostsByIdMutation,
  type Post,
} from '../services/api/api'

export function Home() {
  const { user, logout } = useAuth()
  const [createPost, { isLoading: isCreating }] = usePostPostsMutation()
  const [updatePost] = usePutPostsByIdMutation()
  const [deletePost] = useDeletePostsByIdMutation()
  const { data: posts } = useGetPostsQuery()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [editingPost, setEditingPost] = useState<Post | null>(null)
  const [editTitle, setEditTitle] = useState('')
  const [editContent, setEditContent] = useState('')

  const handleCreatePost = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setSuccess(null)

    try {
      await createPost({
        createPostRequest: { title, content },
      }).unwrap()
      setTitle('')
      setContent('')
      setSuccess('Post created!')
      setTimeout(() => setSuccess(null), 3000)
    } catch {
      setError('Failed to create post')
    }
  }

  const handleEdit = (post: Post) => {
    setEditingPost(post)
    setEditTitle(post.title)
    setEditContent(post.content)
  }

  const handleCancelEdit = () => {
    setEditingPost(null)
    setEditTitle('')
    setEditContent('')
  }

  const handleSaveEdit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!editingPost) return

    try {
      await updatePost({
        id: editingPost.id,
        updatePostRequest: { title: editTitle, content: editContent },
      }).unwrap()
      setEditingPost(null)
    } catch {
      setError('Failed to update post')
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await deletePost({ id }).unwrap()
    } catch {
      setError('Failed to delete post')
    }
  }

  const isOwner = (post: Post) => post.authorId === user.id

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-6 py-4 flex justify-between items-center">
        <h1 className="text-lg font-semibold text-gray-900">no-frame-works</h1>
        <div className="flex items-center gap-4">
          <span className="text-sm text-gray-500">{user.email}</span>
          <button
            onClick={logout}
            className="text-sm text-gray-600 hover:text-gray-900"
          >
            Logout
          </button>
        </div>
      </header>

      <main className="max-w-2xl mx-auto p-6 space-y-8">
        <section className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Create Post</h2>

          <form onSubmit={handleCreatePost} className="space-y-4">
            <div>
              <label htmlFor="title" className="block text-sm font-medium text-gray-700">
                Title
              </label>
              <input
                id="title"
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                required
                className="mt-1 block w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
              />
            </div>

            <div>
              <label htmlFor="content" className="block text-sm font-medium text-gray-700">
                Content
              </label>
              <textarea
                id="content"
                value={content}
                onChange={(e) => setContent(e.target.value)}
                required
                rows={4}
                className="mt-1 block w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none resize-none"
              />
            </div>

            {error && <p className="text-sm text-red-600">{error}</p>}
            {success && <p className="text-sm text-green-600">{success}</p>}

            <button
              type="submit"
              disabled={isCreating}
              className="rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {isCreating ? 'Creating...' : 'Create Post'}
            </button>
          </form>
        </section>

        <section className="space-y-4">
          <h2 className="text-xl font-semibold text-gray-900">Posts</h2>
          {posts && posts.length > 0 ? (
            posts.map((post) => (
              <div key={post.id} className="bg-white rounded-lg shadow p-4">
                {editingPost?.id === post.id ? (
                  <form onSubmit={handleSaveEdit} className="space-y-3">
                    <input
                      type="text"
                      value={editTitle}
                      onChange={(e) => setEditTitle(e.target.value)}
                      required
                      className="block w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                    />
                    <textarea
                      value={editContent}
                      onChange={(e) => setEditContent(e.target.value)}
                      required
                      rows={3}
                      className="block w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none resize-none"
                    />
                    <div className="flex gap-2">
                      <button
                        type="submit"
                        className="rounded bg-blue-600 px-3 py-1 text-sm text-white hover:bg-blue-700"
                      >
                        Save
                      </button>
                      <button
                        type="button"
                        onClick={handleCancelEdit}
                        className="rounded bg-gray-200 px-3 py-1 text-sm text-gray-700 hover:bg-gray-300"
                      >
                        Cancel
                      </button>
                    </div>
                  </form>
                ) : (
                  <>
                    <h3 className="font-medium text-gray-900">{post.title}</h3>
                    <p className="text-sm text-gray-600 mt-1">{post.content}</p>
                    {isOwner(post) && (
                      <div className="flex gap-2 mt-3">
                        <button
                          onClick={() => handleEdit(post)}
                          className="text-xs text-blue-600 hover:underline"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => handleDelete(post.id)}
                          className="text-xs text-red-600 hover:underline"
                        >
                          Delete
                        </button>
                      </div>
                    )}
                  </>
                )}
              </div>
            ))
          ) : (
            <p className="text-sm text-gray-500">No posts yet.</p>
          )}
        </section>
      </main>
    </div>
  )
}
