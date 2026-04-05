import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

export const api = createApi({
  baseQuery: fetchBaseQuery({
    baseUrl: '/api/v1',
    credentials: 'include',
    prepareHeaders: (headers) => {
      headers.set('Accept', 'application/json')
      headers.set('Content-Type', 'application/json')
      return headers
    },
  }),
  reducerPath: 'api',
  refetchOnFocus: true,
  refetchOnReconnect: true,
  endpoints: () => ({}),
})
