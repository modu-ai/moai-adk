---
name: moai-ml-rag
version: 4.0.0
updated: "2025-11-19"
status: stable
stability: stable
description: Retrieval-Augmented Generation systems, vector databases, embedding strategies, and production RAG architectures for enterprise LLM applications. Use when building RAG, semantic search, or knowledge-aware AI systems.
allowed-tools:
  - Read
  - Bash
  - WebSearch
  - WebFetch
---

# Retrieval-Augmented Generation (RAG) — Enterprise

Production-grade RAG systems combining semantic search, vector databases, and LLM generation.

## Quick Start

**5-Minute RAG with LangChain**:

```python
from langchain.vectorstores import Pinecone
from langchain.embeddings import OpenAIEmbeddings
from langchain.llms import OpenAI
from langchain.chains import RetrievalQA

# 1. Create embeddings
embeddings = OpenAIEmbeddings()

# 2. Setup vector store
vectorstore = Pinecone.from_documents(
    documents, embeddings, index_name="docs"
)

# 3. Create RAG chain
qa = RetrievalQA.from_chain_type(
    llm=OpenAI(),
    retriever=vectorstore.as_retriever(search_kwargs={"k": 3})
)

# 4. Ask questions
answer = qa.run("What is the refund policy?")
```

**Auto-triggers**: RAG, retrieval, vector search, semantic search, embedding, knowledge base, Q&A

---

## Core Concepts

### RAG vs Fine-tuning

| Aspect                | RAG                   | Fine-tuning           |
| --------------------- | --------------------- | --------------------- |
| **Knowledge Updates** | Instant (update docs) | Requires retraining   |
| **Cost**              | Low ($50-500/mo)      | High ($1000-10000)    |
| **Time to Deploy**    | Minutes               | Days/weeks            |
| **Knowledge Scope**   | Unlimited             | Fixed (training data) |
| **Hallucinations**    | Reduced (grounded)    | Can still occur       |
| **Use Case**          | QA, search, knowledge | Specialized language  |
| **Recommendation**    | Use first (80% cases) | Only if needed        |

**When to Use**:

- **RAG**: Customer support, FAQ, documentation QA, research
- **Fine-tuning**: Domain jargon mastery, custom behavior
- **Both**: Large knowledge base + custom model behavior

### The 4-Step RAG Pipeline

```
1. INDEXING (Offline)
   → Load documents
   → Split into chunks
   → Create embeddings
   → Store in vector DB

2. RETRIEVAL (Query time)
   → Embed user query
   → Search vector DB
   → Retrieve top-k results

3. RANKING (Optional)
   → Reorder by relevance
   → Filter low-confidence
   → Deduplicate

4. GENERATION (LLM)
   → Build prompt with context
   → Generate answer
   → Post-process output
```

---

## Vector Databases

### Comparison

| Database     | Type        | Best For                 | Scale     | Cost             |
| ------------ | ----------- | ------------------------ | --------- | ---------------- |
| **Pinecone** | Cloud       | Production, ease of use  | Billions  | $70+/mo          |
| **Weaviate** | Open-source | Self-hosted, flexibility | Millions  | Free (self-host) |
| **Milvus**   | Open-source | Large-scale, distributed | Billions  | Free (self-host) |
| **Chroma**   | Local       | Development, prototyping | Thousands | Free             |
| **FAISS**    | Library     | Embedding-only, research | Millions  | Free             |

**Selection Guide**:

- **Development**: Chroma (local, fast setup)
- **Production**: Pinecone (managed, reliable)
- **Self-hosted**: Weaviate (open-source, feature-rich)
- **Large-scale**: Milvus (distributed, scalable)

### Pinecone Example

```python
import pinecone
from sentence_transformers import SentenceTransformer

# Initialize
pinecone.init(api_key="your-key", environment="us-west1-gcp")
index = pinecone.Index("docs")

# Embed and store
model = SentenceTransformer('all-MiniLM-L6-v2')
embedding = model.encode("Python is a programming language")
index.upsert([("doc1", embedding.tolist(), {"text": "..."})])

# Search
results = index.query(embedding.tolist(), top_k=3, include_metadata=True)
```

---

## Embedding Models

### Model Comparison (2025)

| Model                | Dimension | Speed  | Accuracy  | Size  | Use Case         |
| -------------------- | --------- | ------ | --------- | ----- | ---------------- |
| **all-MiniLM-L6-v2** | 384       | Fast   | Good      | 22MB  | Development      |
| **bge-base-en-v1.5** | 768       | Medium | Excellent | 438MB | Production (EN)  |
| **multilingual-e5**  | 768       | Medium | Very Good | 460MB | Multilingual     |
| **OpenAI ada-002**   | 1536      | Fast   | Excellent | API   | Cloud production |
| **Cohere embed-v3**  | 1024      | Fast   | Excellent | API   | Enterprise       |

**Selection**:

- **Development**: all-MiniLM-L6-v2 (fast, small)
- **Production**: bge-large-en-v1.5 or OpenAI ada-002
- **Multilingual**: multilingual-e5-base (100+ languages)
- **Balanced**: bge-base-en-v1.5

### Usage Example

```python
from sentence_transformers import SentenceTransformer

model = SentenceTransformer('all-MiniLM-L6-v2')

# Encode documents
docs = ["Python is great", "Java is popular"]
embeddings = model.encode(docs, batch_size=32)

# Similarity search
query_embedding = model.encode("programming languages")
similarities = model.similarity(query_embedding, embeddings)
```

---

## Chunking Strategies

### Best Practices

**Chunk Size**:

- **Default**: 512 tokens (~400 words)
- **Short answers**: 256 tokens
- **Long context**: 1024 tokens

**Overlap**:

- **Recommended**: 25% overlap (128 tokens for 512-token chunks)
- **Prevents**: Context loss at boundaries

**Example**:

```python
from langchain.text_splitter import RecursiveCharacterTextSplitter

splitter = RecursiveCharacterTextSplitter(
    chunk_size=512,
    chunk_overlap=128,
    separators=["\n\n", "\n", ". ", " ", ""]
)

chunks = splitter.split_text(long_document)
```

---

## RAG Evaluation Metrics

### Key Metrics

| Metric                | Description                                       | Target |
| --------------------- | ------------------------------------------------- | ------ |
| **Hit Rate**          | % queries with relevant doc retrieved             | >90%   |
| **MRR**               | Mean Reciprocal Rank (position of first relevant) | >0.8   |
| **NDCG@k**            | Normalized DCG (ranking quality)                  | >0.85  |
| **Latency**           | Query → Answer time                               | <500ms |
| **Context Relevance** | Retrieved docs actually used in answer            | >80%   |

### Evaluation Example

```python
def calculate_hit_rate(retrieved, relevant):
    """Hit Rate: % of queries where relevant doc was found"""
    hits = len(set(retrieved) & set(relevant))
    return hits / len(relevant) if relevant else 0

def calculate_mrr(retrieved, relevant_doc):
    """Mean Reciprocal Rank: 1/(position of first relevant)"""
    for i, doc in enumerate(retrieved):
        if doc == relevant_doc:
            return 1.0 / (i + 1)
    return 0.0
```

---

## Advanced Patterns

### Hybrid Retrieval (Dense + Sparse)

Combine vector search (semantic) with BM25 (keyword):

```python
from langchain.retrievers import BM25Retriever, EnsembleRetriever

# Dense retrieval
vector_retriever = vectorstore.as_retriever(search_kwargs={"k": 5})

# Sparse retrieval
bm25_retriever = BM25Retriever.from_documents(documents)

# Ensemble (0.5 weight each)
ensemble = EnsembleRetriever(
    retrievers=[vector_retriever, bm25_retriever],
    weights=[0.5, 0.5]
)

results = ensemble.get_relevant_documents("query")
```

### Re-ranking with Cross-Encoders

Improve retrieval quality:

```python
from sentence_transformers import CrossEncoder

# Initial retrieval (fast, lower quality)
initial_results = vectorstore.similarity_search(query, k=20)

# Re-rank (slow, higher quality)
reranker = CrossEncoder('cross-encoder/ms-marco-MiniLM-L-6-v2')
pairs = [[query, doc.page_content] for doc in initial_results]
scores = reranker.predict(pairs)

# Sort and return top-k
reranked = sorted(zip(initial_results, scores), key=lambda x: x[1], reverse=True)
top_results = [doc for doc, score in reranked[:5]]
```

---

## Performance Optimization

### Caching Strategy

```python
import redis
from functools import lru_cache

r = redis.Redis(host='localhost', port=6379, db=0)

def cached_embedding(text):
    """Cache embeddings in Redis"""
    key = f"emb:{hash(text)}"
    cached = r.get(key)

    if cached:
        return pickle.loads(cached)

    embedding = model.encode(text)
    r.setex(key, 86400, pickle.dumps(embedding))  # 24h TTL
    return embedding
```

### Streaming Responses

```python
from langchain.callbacks.streaming_stdout import StreamingStdOutCallbackHandler

qa = RetrievalQA.from_chain_type(
    llm=OpenAI(streaming=True, callbacks=[StreamingStdOutCallbackHandler()]),
    retriever=retriever
)

# Streams tokens as they're generated
qa.run("What is RAG?")
```

---

## Production Best Practices

✅ **DO**:

- Start with simple RAG, iterate
- Evaluate retrieval quality (hit rate, MRR)
- Monitor latency and costs
- Cache embeddings and responses
- Use hybrid retrieval for better coverage
- Implement re-ranking for quality
- Log queries for improvement

❌ **DON'T**:

- Assume retrieval always works
- Skip evaluation metrics
- Use only vector search (add BM25)
- Ignore chunk size tuning
- Store sensitive data unencrypted

---

## Common Issues & Solutions

| Issue                 | Solution                                                          |
| --------------------- | ----------------------------------------------------------------- |
| Low retrieval quality | Try hybrid retrieval, adjust chunk size, use re-ranking           |
| Slow queries          | Cache embeddings, use faster embedding models, optimize vector DB |
| High costs            | Use open-source embeddings, cache responses, batch requests       |
| Hallucinations        | Improve retrieval, use stricter prompts, add confidence scores    |

---

## Advanced Topics

For detailed implementation patterns, see:

- **[examples.md](examples.md)**: Complete RAG implementations, authentication flows, multi-hop reasoning
- **[reference.md](reference.md)**: API references, vector DB setup, production deployment

**Related Skills**:

- `moai-ml-llm-fine-tuning`: LLM fine-tuning patterns
- `moai-domain-ml`: ML best practices
- `moai-essentials-perf`: Performance optimization

---

**Key Libraries**: LangChain 0.2+, LlamaIndex 0.10+, Pinecone 3.0+, sentence-transformers 3.0+

**Version**: 4.0.0  
**Last Updated**: 2025-11-19  
**Status**: Production Ready
