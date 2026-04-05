/** @type {import("@rtk-query/codegen-openapi").ConfigFile} */
const config = {
  schemaFile: '../../../../openapi.yaml',
  apiFile: './client.ts',
  apiImport: 'api',
  outputFile: './api.ts',
  exportName: 'api',
  hooks: { queries: true, lazyQueries: true, mutations: true },
  tag: true,
}

module.exports = config
