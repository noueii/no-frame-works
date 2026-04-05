import { useState } from 'react'
import { useAuth } from '../contexts/AuthContext'

interface Post {
  id: string
  title: string
  content: string
  authorId: string
  authorName: string
}

export function Home() {
  const { sessionToken, logout } = useAuth()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [authorId, setAuthorId] = useState('')
  const [posts, setPosts] = useState<Post[]>([])
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const handleCreatePost = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setSuccess(null)
    setLoading(true)

    try {
      const res = await fetch('/api/v1/posts', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Cookie': `ory_kratos_session=${sessionToken}`,
        },
        credentials: 'include',
        body: JSON.stringify({ title, content, authorId }),
      })

      if (!res.ok) {
        const data = await res.json()
        setError(data.error || 'Failed to create post')
        return
      }

      const post = await res.json()
      setPosts((prev) => [post, ...prev])
      setTitle('')
      setContent('')
      setSuccess('Post created!')
      setTimeout(() => setSuccess(null), 3000)
    } catch {
      setError('Something went wrong')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-6 py-4 flex justify-between items-center">
        <h1 className="text-lg font-semibold text-gray-900">no-frame-works</h1>
        <button
          onClick={logout}
          className="text-sm text-gray-600 hover:text-gray-900"
        >
          Logout
        </button>
      </header>

      <main className="max-w-2xl mx-auto p-6 space-y-8">
        <section className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Create Post</h2>

          <form onSubmit={handleCreatePost} className="space-y-4">
            <div>
              <label htmlFor="authorId" className="block text-sm font-medium text-gray-700">
                Author ID
              </label>
              <input
                id="authorId"
                type="text"
                value={authorId}
                onChange={(e) => setAuthorId(e.target.value)}
                required
                placeholder="UUID of the author"
                className="mt-1 block w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
              />
            </div>

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
              disabled={loading}
              className="rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? 'Creating...' : 'Create Post'}
            </button>
          </form>
        </section>

        {posts.length > 0 && (
          <section className="space-y-4">
            <h2 className="text-xl font-semibold text-gray-900">Posts</h2>
            {posts.map((post) => (
              <div key={post.id} className="bg-white rounded-lg shadow p-4">
                <h3 className="font-medium text-gray-900">{post.title}</h3>
                <p className="text-sm text-gray-600 mt-1">{post.content}</p>
                <p className="text-xs text-gray-400 mt-2">by {post.authorName}</p>
              </div>
            ))}
          </section>
        )}
      </main>
    </div>
  )
}
