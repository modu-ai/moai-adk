---
name: storage-cdn
description: File storage with image transformations and CDN delivery
parent-skill: moai-platform-supabase
version: 1.0.0
updated: 2026-01-06
---

# Storage and CDN Module

## Overview

Supabase Storage provides file storage with automatic image transformations, CDN delivery, and fine-grained access control through storage policies.

## Basic Upload

### Upload File

```typescript
const supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY)

async function uploadFile(file: File, bucket: string, path: string) {
  const { data, error } = await supabase.storage
    .from(bucket)
    .upload(path, file, {
      cacheControl: '3600',
      upsert: false
    })

  if (error) throw error
  return data.path
}
```

### Upload with User Context

```typescript
async function uploadUserFile(file: File, userId: string) {
  const fileName = `${userId}/${Date.now()}-${file.name}`

  const { data, error } = await supabase.storage
    .from('user-files')
    .upload(fileName, file, {
      cacheControl: '3600',
      upsert: false
    })

  if (error) throw error
  return data
}
```

## Image Transformations

### Get Transformed URL

```typescript
async function uploadImage(file: File, userId: string) {
  const fileName = `${userId}/${Date.now()}-${file.name}`

  const { data, error } = await supabase.storage
    .from('images')
    .upload(fileName, file, { cacheControl: '3600', upsert: false })

  if (error) throw error

  // Get original URL
  const { data: { publicUrl } } = supabase.storage
    .from('images')
    .getPublicUrl(fileName)

  // Get resized URL
  const { data: { publicUrl: resizedUrl } } = supabase.storage
    .from('images')
    .getPublicUrl(fileName, {
      transform: { width: 800, height: 600, resize: 'contain' }
    })

  // Get thumbnail URL
  const { data: { publicUrl: thumbnailUrl } } = supabase.storage
    .from('images')
    .getPublicUrl(fileName, {
      transform: { width: 200, height: 200, resize: 'cover' }
    })

  return { originalPath: data.path, publicUrl, resizedUrl, thumbnailUrl }
}
```

### Transform Options

Available transformation parameters:

- width: Target width in pixels
- height: Target height in pixels
- resize: 'cover' | 'contain' | 'fill'
- format: 'origin' | 'avif' | 'webp'
- quality: 1-100

### Example Transforms

```typescript
// Square thumbnail with crop
const thumbnail = supabase.storage
  .from('images')
  .getPublicUrl(path, {
    transform: { width: 150, height: 150, resize: 'cover' }
  })

// WebP format for smaller size
const webp = supabase.storage
  .from('images')
  .getPublicUrl(path, {
    transform: { width: 800, format: 'webp', quality: 80 }
  })

// Responsive image
const responsive = supabase.storage
  .from('images')
  .getPublicUrl(path, {
    transform: { width: 400, resize: 'contain' }
  })
```

## Bucket Management

### Create Bucket

```sql
-- Via SQL
INSERT INTO storage.buckets (id, name, public)
VALUES ('images', 'images', true);
```

### Bucket Policies

```sql
-- Allow authenticated users to upload to their folder
CREATE POLICY "User upload" ON storage.objects
FOR INSERT
TO authenticated
WITH CHECK (bucket_id = 'user-files' AND (storage.foldername(name))[1] = auth.uid()::text);

-- Allow public read on images bucket
CREATE POLICY "Public read" ON storage.objects
FOR SELECT
TO public
USING (bucket_id = 'images');

-- Allow users to delete their own files
CREATE POLICY "User delete" ON storage.objects
FOR DELETE
TO authenticated
USING (bucket_id = 'user-files' AND (storage.foldername(name))[1] = auth.uid()::text);
```

## Download Files

### Get Signed URL

For private buckets:

```typescript
async function getSignedUrl(bucket: string, path: string, expiresIn: number = 3600) {
  const { data, error } = await supabase.storage
    .from(bucket)
    .createSignedUrl(path, expiresIn)

  if (error) throw error
  return data.signedUrl
}
```

### Download File

```typescript
async function downloadFile(bucket: string, path: string) {
  const { data, error } = await supabase.storage
    .from(bucket)
    .download(path)

  if (error) throw error
  return data // Blob
}
```

## File Management

### List Files

```typescript
async function listFiles(bucket: string, folder: string) {
  const { data, error } = await supabase.storage
    .from(bucket)
    .list(folder, {
      limit: 100,
      offset: 0,
      sortBy: { column: 'created_at', order: 'desc' }
    })

  if (error) throw error
  return data
}
```

### Delete File

```typescript
async function deleteFile(bucket: string, paths: string[]) {
  const { data, error } = await supabase.storage
    .from(bucket)
    .remove(paths)

  if (error) throw error
  return data
}
```

### Move/Rename File

```typescript
async function moveFile(bucket: string, fromPath: string, toPath: string) {
  const { data, error } = await supabase.storage
    .from(bucket)
    .move(fromPath, toPath)

  if (error) throw error
  return data
}
```

## React Integration

### Upload Component

```typescript
function FileUploader({ bucket, onUpload }: Props) {
  const [uploading, setUploading] = useState(false)

  async function handleUpload(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    if (!file) return

    setUploading(true)
    try {
      const path = `${Date.now()}-${file.name}`
      const { error } = await supabase.storage
        .from(bucket)
        .upload(path, file)

      if (error) throw error
      onUpload(path)
    } finally {
      setUploading(false)
    }
  }

  return (
    <input
      type="file"
      onChange={handleUpload}
      disabled={uploading}
    />
  )
}
```

### Image with Fallback

```typescript
function StorageImage({ path, bucket, width, height, fallback }: Props) {
  const { data: { publicUrl } } = supabase.storage
    .from(bucket)
    .getPublicUrl(path, {
      transform: { width, height, resize: 'cover' }
    })

  return (
    <img
      src={publicUrl}
      alt=""
      onError={(e) => { e.currentTarget.src = fallback }}
      width={width}
      height={height}
    />
  )
}
```

## Best Practices

File Organization:
- Use user ID as folder prefix for user content
- Include timestamp in filenames to prevent collisions
- Use consistent naming conventions

Performance:
- Set appropriate cache-control headers
- Use image transformations instead of storing multiple sizes
- Leverage CDN for global delivery

Security:
- Always use RLS-style policies for storage
- Use signed URLs for private content
- Validate file types before upload

## Context7 Query Examples

For latest Storage documentation:

Topic: "supabase storage upload download"
Topic: "storage image transformations"
Topic: "storage bucket policies"

---

Related Modules:
- row-level-security.md - Storage access policies
- typescript-patterns.md - Client patterns
