# Web API - Working Examples

> Real-world REST and GraphQL API implementations

---

## Example 1: Express REST API

### server.ts
```typescript
import express from 'express';

const app = express();
app.use(express.json());

// GET all users
app.get('/api/users', async (req, res) => {
  const users = await db.getAllUsers();
  res.json(users);
});

// POST create user
app.post('/api/users', async (req, res) => {
  const user = await db.createUser(req.body);
  res.status(201).json(user);
});

app.listen(3000, () => console.log('Server running on port 3000'));
```

---

## Example 2: GraphQL API

### schema.ts
```typescript
import { ApolloServer, gql } from 'apollo-server';

const typeDefs = gql`
  type User {
    id: ID!
    name: String!
    email: String!
  }

  type Query {
    users: [User!]!
    user(id: ID!): User
  }
`;

const resolvers = {
  Query: {
    users: () => db.getAllUsers(),
    user: (_: any, { id }: { id: string }) => db.findUser(id),
  },
};

const server = new ApolloServer({ typeDefs, resolvers });
server.listen().then(({ url }) => console.log(`Server ready at ${url}`));
```

---

**Last Updated**: 2025-10-22
**Technologies**: Express, Apollo Server, GraphQL
