FROM node:alpine
LABEL description="Instant high-performance GraphQL API for your PostgreSQL database https://github.com/graphile/postgraphile"

WORKDIR /graphql
COPY . /graphql

RUN cd exec-sql && npm i

# Install PostGraphile and PostGraphile connection filter plugin
RUN npm install -g postgraphile
RUN npm install -g postgraphile-plugin-connection-filter

RUN npm install -g @graphile-contrib/pg-simplify-inflector

EXPOSE 5433
ENTRYPOINT ["postgraphile", "-n", "0.0.0.0"]
