# moai-security-secrets: Production Examples

## Example 1: Vault Integration

```javascript
const vault = require('@hashicorp/vault-client');

const client = new vault.ApiClient({
  endpoint: process.env.VAULT_ADDR,
  token: process.env.VAULT_TOKEN
});

async function getSecret(path) {
  const response = await client.read(`secret/data/${path}`);
  return response.data.data;
}

async function rotateSecret(path) {
  const newSecret = crypto.randomBytes(32).toString('hex');
  await client.write(`secret/data/${path}`, { data: { value: newSecret } });
}

const dbPassword = await getSecret('database/password');
```

## Example 2: Secrets Rotation

```javascript
const cron = require('node-cron');

cron.schedule('0 2 * * 0', async () => {
  console.log('Rotating secrets...');
  
  await rotateSecret('database/password');
  await rotateSecret('api/key');
  
  console.log('Secrets rotated successfully');
});
```

