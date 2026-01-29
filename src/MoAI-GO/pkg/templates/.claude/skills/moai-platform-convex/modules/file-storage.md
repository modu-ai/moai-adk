# File Storage Module

File upload, storage, and serving patterns for Convex applications.

---

## Storage Fundamentals

### Generate Upload URL

```typescript
// convex/functions/files.ts
import { mutation, query } from '../_generated/server'
import { v } from 'convex/values'

export const generateUploadUrl = mutation({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new Error('Unauthorized')
    
    return await ctx.storage.generateUploadUrl()
  }
})
```

### Save File Reference

```typescript
export const saveFile = mutation({
  args: {
    storageId: v.id('_storage'),
    fileName: v.string(),
    fileType: v.string(),
    fileSize: v.number()
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new Error('Unauthorized')
    
    return await ctx.db.insert('files', {
      storageId: args.storageId,
      fileName: args.fileName,
      fileType: args.fileType,
      fileSize: args.fileSize,
      uploaderId: identity.subject,
      uploadedAt: Date.now()
    })
  }
})
```

### Get File URL

```typescript
export const getFileUrl = query({
  args: { storageId: v.id('_storage') },
  handler: async (ctx, args) => {
    return await ctx.storage.getUrl(args.storageId)
  }
})
```

---

## File Schema

```typescript
// convex/schema.ts
import { defineSchema, defineTable } from 'convex/server'
import { v } from 'convex/values'

export default defineSchema({
  files: defineTable({
    storageId: v.id('_storage'),
    fileName: v.string(),
    fileType: v.string(),
    fileSize: v.number(),
    uploaderId: v.string(),
    uploadedAt: v.number(),
    // Optional metadata
    description: v.optional(v.string()),
    tags: v.optional(v.array(v.string())),
    isPublic: v.optional(v.boolean())
  })
    .index('by_uploader', ['uploaderId'])
    .index('by_type', ['fileType'])
    .index('by_uploaded', ['uploadedAt'])
})
```

---

## Client-Side Upload

### Basic File Upload Hook

```typescript
// src/hooks/useFileUpload.ts
import { useMutation } from 'convex/react'
import { api } from '../../convex/_generated/api'

export function useFileUpload() {
  const generateUploadUrl = useMutation(api.functions.files.generateUploadUrl)
  const saveFile = useMutation(api.functions.files.saveFile)
  
  return async (file: File) => {
    // Get upload URL from Convex
    const uploadUrl = await generateUploadUrl()
    
    // Upload file to Convex storage
    const response = await fetch(uploadUrl, {
      method: 'POST',
      headers: { 'Content-Type': file.type },
      body: file
    })
    
    if (!response.ok) {
      throw new Error('Upload failed')
    }
    
    const { storageId } = await response.json()
    
    // Save file reference in database
    const fileId = await saveFile({
      storageId,
      fileName: file.name,
      fileType: file.type,
      fileSize: file.size
    })
    
    return { fileId, storageId }
  }
}
```

### Upload with Progress

```typescript
export function useFileUploadWithProgress() {
  const generateUploadUrl = useMutation(api.functions.files.generateUploadUrl)
  const saveFile = useMutation(api.functions.files.saveFile)
  
  return async (
    file: File,
    onProgress: (progress: number) => void
  ) => {
    const uploadUrl = await generateUploadUrl()
    
    return new Promise<{ fileId: string; storageId: string }>(
      (resolve, reject) => {
        const xhr = new XMLHttpRequest()
        
        xhr.upload.addEventListener('progress', (event) => {
          if (event.lengthComputable) {
            const progress = (event.loaded / event.total) * 100
            onProgress(progress)
          }
        })
        
        xhr.addEventListener('load', async () => {
          if (xhr.status === 200) {
            const { storageId } = JSON.parse(xhr.responseText)
            const fileId = await saveFile({
              storageId,
              fileName: file.name,
              fileType: file.type,
              fileSize: file.size
            })
            resolve({ fileId, storageId })
          } else {
            reject(new Error('Upload failed'))
          }
        })
        
        xhr.addEventListener('error', () => reject(new Error('Upload failed')))
        
        xhr.open('POST', uploadUrl)
        xhr.setRequestHeader('Content-Type', file.type)
        xhr.send(file)
      }
    )
  }
}
```

---

## File Upload Component

### Drag and Drop Upload

```typescript
// src/components/FileUploader.tsx
import { useFileUpload } from '../hooks/useFileUpload'
import { useState, useCallback } from 'react'

export function FileUploader({ onUploadComplete }: {
  onUploadComplete: (fileId: string) => void
}) {
  const upload = useFileUpload()
  const [isDragging, setIsDragging] = useState(false)
  const [isUploading, setIsUploading] = useState(false)
  
  const handleDrop = useCallback(async (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
    
    const files = Array.from(e.dataTransfer.files)
    if (files.length === 0) return
    
    setIsUploading(true)
    try {
      const { fileId } = await upload(files[0])
      onUploadComplete(fileId)
    } finally {
      setIsUploading(false)
    }
  }, [upload, onUploadComplete])
  
  return (
    <div
      onDragOver={(e) => { e.preventDefault(); setIsDragging(true) }}
      onDragLeave={() => setIsDragging(false)}
      onDrop={handleDrop}
      className={`upload-zone ${isDragging ? 'dragging' : ''}`}
    >
      {isUploading ? 'Uploading...' : 'Drop files here or click to upload'}
    </div>
  )
}
```

### Multiple File Upload

```typescript
export function MultiFileUploader({ onUploadComplete }: {
  onUploadComplete: (fileIds: string[]) => void
}) {
  const upload = useFileUpload()
  const [uploading, setUploading] = useState<string[]>([])
  const [completed, setCompleted] = useState<string[]>([])
  
  const handleFiles = async (files: FileList) => {
    const fileArray = Array.from(files)
    setUploading(fileArray.map(f => f.name))
    
    const results = await Promise.all(
      fileArray.map(async (file) => {
        const { fileId } = await upload(file)
        setCompleted(prev => [...prev, file.name])
        return fileId
      })
    )
    
    setUploading([])
    setCompleted([])
    onUploadComplete(results)
  }
  
  return (
    <div>
      <input
        type="file"
        multiple
        onChange={(e) => e.target.files && handleFiles(e.target.files)}
      />
      {uploading.length > 0 && (
        <div>
          Uploading: {uploading.length - completed.length} remaining
        </div>
      )}
    </div>
  )
}
```

---

## File Display

### Image Display Component

```typescript
// src/components/StorageImage.tsx
import { useQuery } from 'convex/react'
import { api } from '../../convex/_generated/api'
import { Id } from '../../convex/_generated/dataModel'

export function StorageImage({ storageId, alt }: {
  storageId: Id<'_storage'>
  alt: string
}) {
  const url = useQuery(api.functions.files.getFileUrl, { storageId })
  
  if (!url) return <ImagePlaceholder />
  
  return <img src={url} alt={alt} />
}
```

### File List with Download

```typescript
export function FileList({ uploaderId }: { uploaderId: string }) {
  const files = useQuery(api.functions.files.listByUploader, { uploaderId })
  
  if (files === undefined) return <Loading />
  
  return (
    <ul>
      {files.map((file) => (
        <FileListItem key={file._id} file={file} />
      ))}
    </ul>
  )
}

function FileListItem({ file }: { file: FileRecord }) {
  const url = useQuery(api.functions.files.getFileUrl, {
    storageId: file.storageId
  })
  
  return (
    <li>
      <span>{file.fileName}</span>
      <span>{formatFileSize(file.fileSize)}</span>
      {url && (
        <a href={url} download={file.fileName}>Download</a>
      )}
    </li>
  )
}
```

---

## File Management

### Delete File

```typescript
// convex/functions/files.ts
export const deleteFile = mutation({
  args: { fileId: v.id('files') },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new Error('Unauthorized')
    
    const file = await ctx.db.get(args.fileId)
    if (!file) throw new Error('File not found')
    
    if (file.uploaderId !== identity.subject) {
      throw new Error('Forbidden')
    }
    
    // Delete from storage
    await ctx.storage.delete(file.storageId)
    
    // Delete database record
    await ctx.db.delete(args.fileId)
  }
})
```

### List Files by Uploader

```typescript
export const listByUploader = query({
  args: { uploaderId: v.string() },
  handler: async (ctx, args) => {
    const files = await ctx.db
      .query('files')
      .withIndex('by_uploader', (q) => q.eq('uploaderId', args.uploaderId))
      .order('desc')
      .collect()
    
    return files
  }
})
```

### Get File with URL

```typescript
export const getFileWithUrl = query({
  args: { fileId: v.id('files') },
  handler: async (ctx, args) => {
    const file = await ctx.db.get(args.fileId)
    if (!file) return null
    
    const url = await ctx.storage.getUrl(file.storageId)
    
    return { ...file, url }
  }
})
```

---

## Image Processing

### Image with Thumbnail

```typescript
// Schema with thumbnail
files: defineTable({
  storageId: v.id('_storage'),
  thumbnailStorageId: v.optional(v.id('_storage')),
  fileName: v.string(),
  fileType: v.string(),
  fileSize: v.number(),
  uploaderId: v.string(),
  uploadedAt: v.number()
})

// Action to generate thumbnail
export const generateThumbnail = action({
  args: { fileId: v.id('files') },
  handler: async (ctx, args) => {
    const file = await ctx.runQuery(internal.files.getById, { id: args.fileId })
    if (!file) throw new Error('File not found')
    
    const originalUrl = await ctx.storage.getUrl(file.storageId)
    if (!originalUrl) throw new Error('File URL not found')
    
    // Call external image processing service
    const thumbnailBlob = await generateThumbnailBlob(originalUrl)
    
    // Upload thumbnail
    const uploadUrl = await ctx.storage.generateUploadUrl()
    await fetch(uploadUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'image/jpeg' },
      body: thumbnailBlob
    })
    
    const { storageId: thumbnailStorageId } = await response.json()
    
    // Update file record
    await ctx.runMutation(internal.files.updateThumbnail, {
      fileId: args.fileId,
      thumbnailStorageId
    })
  }
})
```

---

## Best Practices

Upload Handling:
- Always validate file types before upload
- Set maximum file size limits
- Use progress indicators for large files
- Handle upload failures gracefully

Storage Management:
- Delete storage files when deleting records
- Use appropriate content types
- Consider thumbnail generation for images
- Implement file cleanup for orphaned storage

Security:
- Validate user permissions before operations
- Use signed URLs for sensitive files
- Don't expose internal storage IDs directly
- Implement rate limiting for uploads

---

Version: 1.0.0
Module: File Storage
Parent Skill: moai-platform-convex
