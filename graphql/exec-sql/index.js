const { makeExtendSchemaPlugin, gql } = require("graphile-utils");

module.exports = makeExtendSchemaPlugin({
  typeDefs: gql`
    extend type Query {
      execSQL(query: String!): JSON
    }
  `,
  resolvers: {
    Query: {
      async execSQL(_parent, args, context, _info) {
        const { rows } = await context.pgClient.query(args.query)
        return rows;
      },
    },
  },
});
