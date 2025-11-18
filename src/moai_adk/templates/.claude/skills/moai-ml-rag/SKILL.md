---
name: moai-ml-rag
version: 4.0.0
updated: '2025-11-19'
status: stable
stability: stable
description: Retrieval-Augmented Generation systems, vector databases, embedding strategies, and production RAG architectures for enterprise LLM applications
allowed-tools:
- Read
- Bash
- WebSearch
- WebFetch
---

# Retrieval-Augmented Generation (RAG) — Enterprise  

## Quick Summary

**Primary Focus**: Production-grade Retrieval-Augmented Generation (RAG) systems combining semantic search, vector databases, and LLM generation with enterprise quality validation

**Best For**: Building knowledge-aware AI systems, reducing hallucinations, maintaining data freshness, domain-specific LLM applications, question-answering systems

**Key Libraries**: LangChain 0.2+, LlamaIndex 0.10+, Pinecone 3.0+, Weaviate latest, sentence-transformers 3.0+, Cohere 5.x

**Key VectorDBs**: Pinecone (cloud), Weaviate (open-source), Milvus (large-scale), Chroma (local), FAISS (embedding-only)

**Auto-triggers**: RAG, retrieval, vector search, semantic search, embedding, knowledge base, Q&A, LLM context, generation, knowledge-aware

| Component | Version | Release | Support |
|-----------|---------|---------|---------|
| LangChain | 0.2.0+ | 2025-10 | Active |
| LlamaIndex | 0.10.0+ | 2025-09 | Active |
| Pinecone | 3.0.0+ | 2025-08 | Active |
| Weaviate | 1.26+ | 2025-11 | Active |
| sentence-transformers | 3.0.0+ | 2025-10 | Active |
| Milvus | 2.4.0+ | 2025-10 | Active |

---

## Three-Level Learning Path

### Level 1: Fundamentals (Read examples.md)

Core RAG concepts with practical examples:

- **RAG vs Fine-tuning**: When to use retrieval vs adaptation
- **4-Step RAG Pipeline**: Indexing, retrieval, ranking, generation
- **Embedding Basics**: Vector representations, similarity metrics
- **Vector Database Comparison**: Cloud vs open-source tradeoffs
- **Chunking Strategies**: Document splitting, overlap management
- **Evaluation Metrics**: Hit rate, MRR, NDCG for RAG quality
- **Examples**: See `examples.md` for complete code samples

### Level 2: Advanced Patterns (See reference.md)

Production-ready enterprise patterns:

- **Hybrid Retrieval**: Dense (vector) + sparse (BM25) combination
- **Re-ranking Strategies**: Improving result quality with cross-encoders
- **Multi-hop Reasoning**: Complex question decomposition
- **Caching Optimization**: Redis for embeddings and responses
- **Streaming**: Token-by-token generation for responsive UX
- **Monitoring**: Search quality, latency, hallucination tracking
- **Pattern Reference**: See `reference.md` for enterprise patterns

### Level 3: Production Deployment (Consult infrastructure skills)

Enterprise deployment and optimization:

- **Distributed Search**: Scaling across multiple vector indices
- **Cost Optimization**: Vector DB pricing, embedding caching
- **Security**: PII filtering, access control, audit logging
- **Observability**: Metrics, tracing, alerting for RAG systems
- **A/B Testing**: Comparing retrieval strategies, model versions
- **Details**: Skill("moai-essentials-perf"), Skill("moai-security-backend")

---

## 1. RAG Fundamentals

### What is RAG?

Retrieval-Augmented Generation (RAG) enhances language models by providing external context:

```
Traditional LLM:
Question → Model (trained knowledge only) → Answer
Problem: Limited knowledge, hallucinations, outdated info

RAG System:
Question → Retrieve relevant documents → Provide as context → Model → Answer
Benefit: Current information, factual grounding, source attribution
```

### Why RAG Over Fine-tuning?

| Aspect | RAG | Fine-tuning |
|--------|-----|------------|
| **Knowledge Updates** | Instant (update documents) | Requires retraining |
| **Cost** | Low (semantic search) | High (GPU training) |
| **Time to Deploy** | Minutes | Days/weeks |
| **Knowledge Scope** | Unlimited (any documents) | Fixed (training data) |
| **Accuracy** | Good (with quality retrieval) | Excellent (if trained well) |
| **Hallucinations** | Reduced (grounded in docs) | Can still occur |
| **Use Case** | QA, search, knowledge base | Specialized language adaptation |
| **Recommendation** | Use first (80% of cases) | Use only if fine-tuning needed |

**When to Use Each**:
- RAG: Customer support, FAQ, knowledge base, research, documentation QA
- Fine-tuning: Specialized language style, domain jargon mastery, custom behavior
- Both: Large knowledge base + custom model behavior

#### Example 1: RAG vs Fine-tuning Architecture

```python
# rag_vs_finetuning_comparison.py
from typing import List
import numpy as np

class RAGSystem:
    """Retrieval-Augmented Generation approach"""
    
    def __init__(self, vector_db, llm):
        self.vector_db = vector_db  # Pinecone, Weaviate, etc.
        self.llm = llm
    
    def answer_question(self, question: str, top_k: int = 3) -> str:
        """RAG pipeline: Retrieve → Rank → Generate"""
        
        # Step 1: Retrieve relevant documents
        documents = self.vector_db.search(
            query=question,
            top_k=top_k
        )
        
        # Step 2: Rank by relevance (optional)
        ranked_docs = self.rank_documents(documents)
        
        # Step 3: Build prompt with context
        context = "\n".join([doc.text for doc in ranked_docs])
        prompt = f"""Answer the question based on the context provided.
        
Context:
{context}

Question: {question}

Answer:"""
        
        # Step 4: Generate answer
        answer = self.llm.generate(prompt)
        
        return answer
    
    def update_knowledge(self, new_documents: List[str]):
        """Easy knowledge update - just add new documents"""
        for doc in new_documents:
            embedding = self.vector_db.embed(doc)
            self.vector_db.add(embedding, doc)


class FineTunedSystem:
    """Fine-tuning approach"""
    
    def __init__(self, model_path: str):
        self.model = self.load_model(model_path)
    
    def answer_question(self, question: str) -> str:
        """Direct generation without retrieval"""
        
        # Model has specialized knowledge from training
        answer = self.model.generate(question)
        
        return answer
    
    def update_knowledge(self, new_examples: List[dict]):
        """Knowledge update requires retraining"""
        
        print("Retraining required - expensive and time-consuming")
        # 1. Add new_examples to training data
        # 2. Retrain model (hours/days on GPU)
        # 3. Evaluate performance
        # 4. Deploy new version
        # 5. Monitor in production


# Comparison table
comparison = {
    "Knowledge Update": {
        "RAG": "5 minutes (add documents + embed)",
        "Fine-tuning": "1-3 days (retrain + validate + deploy)"
    },
    "System Cost": {
        "RAG": "$50-500/month (vector DB)",
        "Fine-tuning": "$1000-10000 (GPU hours)"
    },
    "Answer Accuracy": {
        "RAG": "Good (if retrieval works)",
        "Fine-tuning": "Excellent (for trained domain)"
    },
    "Hallucinations": {
        "RAG": "Lower (grounded in docs)",
        "Fine-tuning": "Can still occur"
    },
}

print("RAG vs Fine-tuning Comparison:")
print("=" * 70)
for aspect, comparison_data in comparison.items():
    print(f"\n{aspect}:")
    print(f"  RAG:          {comparison_data['RAG']}")
    print(f"  Fine-tuning:  {comparison_data['Fine-tuning']}")


# Hybrid approach (recommended for production)
class HybridSystem:
    """Combine RAG + Fine-tuning for best results"""
    
    def __init__(self, fine_tuned_model, vector_db):
        self.model = fine_tuned_model  # Specialized language style
        self.vector_db = vector_db      # Current knowledge
    
    def answer_question(self, question: str) -> str:
        """RAG + Fine-tuned model = Best of both"""
        
        # Retrieve context
        documents = self.vector_db.search(question, top_k=3)
        context = "\n".join([doc.text for doc in documents])
        
        # Use fine-tuned model with current context
        prompt = f"Context:\n{context}\n\nQuestion: {question}"
        answer = self.model.generate(prompt)
        
        return answer
    
    def update_knowledge(self, new_documents: List[str]):
        """Quick updates via RAG, periodic fine-tuning"""
        # Add to vector DB (instant)
        for doc in new_documents:
            self.vector_db.add_document(doc)
        
        # Periodically fine-tune on new examples (weekly)
        # if new_documents count > THRESHOLD:
        #     retrain_model()
```

### The 4-Step RAG Pipeline

RAG follows a consistent 4-step process:

**Step 1: Indexing** (Offline preparation)
- Load documents
- Split into chunks
- Create embeddings
- Store in vector database

**Step 2: Retrieval** (At query time)
- Embed user query
- Search vector database
- Retrieve top-k similar documents
- Return ranked results

**Step 3: Ranking** (Optional improvement)
- Reorder results by relevance
- Filter low-confidence matches
- Deduplicate overlapping docs
- Apply domain-specific scoring

**Step 4: Generation** (LLM response)
- Build prompt with context
- Send to LLM
- Stream or batch generation
- Post-process output

#### Example 2: Complete RAG Pipeline Implementation

```python
# complete_rag_pipeline.py
from typing import List, Tuple
import numpy as np
from dataclasses import dataclass

@dataclass
class Document:
    content: str
    source: str
    metadata: dict

class RAGPipeline:
    """Complete 4-step RAG implementation"""
    
    def __init__(self, embedding_model, vector_db, llm):
        self.embedding_model = embedding_model
        self.vector_db = vector_db
        self.llm = llm
    
    # STEP 1: INDEXING (Offline)
    def index_documents(self, documents: List[Document], chunk_size: int = 512):
        """Index documents into vector database"""
        
        print(f"Indexing {len(documents)} documents...")
        
        for doc in documents:
            # Split into chunks
            chunks = self.chunk_document(doc.content, chunk_size)
            
            for i, chunk in enumerate(chunks):
                # Create embedding
                embedding = self.embedding_model.encode(chunk)
                
                # Store in vector DB
                self.vector_db.add(
                    id=f"{doc.source}_{i}",
                    vector=embedding,
                    metadata={
                        "source": doc.source,
                        "chunk_index": i,
                        "content": chunk,
                        **doc.metadata
                    }
                )
        
        print(f"Indexed {len(documents)} documents successfully")
    
    def chunk_document(self, text: str, chunk_size: int = 512) -> List[str]:
        """Split document into overlapping chunks"""
        
        # Simple chunking (in production use LangChain/LlamaIndex)
        chunks = []
        overlap = chunk_size // 4
        
        words = text.split()
        
        for i in range(0, len(words), chunk_size - overlap):
            chunk = " ".join(words[i:i + chunk_size])
            chunks.append(chunk)
        
        return chunks
    
    # STEP 2: RETRIEVAL (Query time)
    def retrieve(self, query: str, top_k: int = 5) -> List[Document]:
        """Retrieve relevant documents"""
        
        # Embed query
        query_embedding = self.embedding_model.encode(query)
        
        # Search vector DB
        results = self.vector_db.search(
            vector=query_embedding,
            top_k=top_k
        )
        
        # Convert to Document objects
        documents = [
            Document(
                content=result['metadata']['content'],
                source=result['metadata']['source'],
                metadata=result['metadata']
            )
            for result in results
        ]
        
        return documents
    
    # STEP 3: RANKING (Improve quality)
    def rank_documents(self, documents: List[Document], query: str) -> List[Document]:
        """Re-rank documents by relevance"""
        
        # Simple scoring (in production use cross-encoders)
        scores = []
        
        for doc in documents:
            # Calculate relevance score
            query_terms = set(query.lower().split())
            doc_terms = set(doc.content.lower().split())
            
            # Jaccard similarity
            intersection = len(query_terms & doc_terms)
            union = len(query_terms | doc_terms)
            score = intersection / union if union > 0 else 0
            
            scores.append((doc, score))
        
        # Sort by score
        documents_ranked = [doc for doc, score in sorted(
            scores,
            key=lambda x: x[1],
            reverse=True
        )]
        
        return documents_ranked
    
    # STEP 4: GENERATION (LLM response)
    def generate_answer(self, query: str, documents: List[Document]) -> str:
        """Generate answer using LLM with context"""
        
        # Build context
        context = "\n\n".join([
            f"[{doc.source}]\n{doc.content}"
            for doc in documents
        ])
        
        # Build prompt
        prompt = f"""You are a helpful assistant. Answer the question based on the provided context.

Context:
{context}

Question: {query}

Answer:"""
        
        # Generate with LLM
        answer = self.llm.generate(
            prompt=prompt,
            temperature=0.7,
            max_tokens=500
        )
        
        return answer
    
    # COMPLETE RAG PIPELINE
    def answer_question(self, query: str, top_k: int = 5) -> Tuple[str, List[Document]]:
        """Complete RAG pipeline"""
        
        print(f"\nProcessing query: {query}")
        
        # Step 2: Retrieve
        print("Step 2: Retrieving documents...")
        documents = self.retrieve(query, top_k)
        
        # Step 3: Rank
        print("Step 3: Ranking documents...")
        documents = self.rank_documents(documents, query)
        
        # Step 4: Generate
        print("Step 4: Generating answer...")
        answer = self.generate_answer(query, documents)
        
        return answer, documents


# Quality evaluation metrics
class RAGEvaluator:
    """Evaluate RAG system quality"""
    
    @staticmethod
    def calculate_hit_rate(retrieved_docs: List[str], relevant_docs: List[str]) -> float:
        """Hit Rate: % of queries where relevant doc was retrieved"""
        
        retrieved_set = set(retrieved_docs)
        relevant_set = set(relevant_docs)
        
        if not relevant_set:
            return 0.0
        
        hits = len(retrieved_set & relevant_set)
        return hits / len(relevant_set)
    
    @staticmethod
    def calculate_mrr(retrieved_docs: List[str], relevant_doc: str) -> float:
        """Mean Reciprocal Rank: Position of first relevant doc"""
        
        for i, doc in enumerate(retrieved_docs):
            if doc == relevant_doc:
                return 1.0 / (i + 1)
        
        return 0.0  # Relevant doc not found
    
    @staticmethod
    def calculate_ndcg(retrieved_docs: List[str], relevant_docs: List[str], k: int = 10) -> float:
        """Normalized DCG: Quality of ranking"""
        
        # DCG calculation
        dcg = 0.0
        for i, doc in enumerate(retrieved_docs[:k]):
            relevance = 1.0 if doc in relevant_docs else 0.0
            dcg += relevance / np.log2(i + 2)
        
        # Ideal DCG
        idcg = 0.0
        for i in range(min(len(relevant_docs), k)):
            idcg += 1.0 / np.log2(i + 2)
        
        # Normalize
        ndcg = dcg / idcg if idcg > 0 else 0.0
        
        return ndcg


# Example usage
if __name__ == "__main__":
    # Simulated components
    class DummyEmbeddingModel:
        def encode(self, text: str) -> np.ndarray:
            return np.random.randn(384)
    
    class DummyVectorDB:
        def add(self, **kwargs): pass
        def search(self, **kwargs): return []
    
    class DummyLLM:
        def generate(self, prompt: str, **kwargs) -> str:
            return "Sample answer based on retrieved context."
    
    # Initialize RAG
    rag = RAGPipeline(
        embedding_model=DummyEmbeddingModel(),
        vector_db=DummyVectorDB(),
        llm=DummyLLM()
    )
    
    # Index documents
    documents = [
        Document(
            content="Python is a high-level programming language.",
            source="docs.md",
            metadata={"type": "documentation"}
        )
    ]
    
    rag.index_documents(documents)
    
    # Answer question
    answer, sources = rag.answer_question("What is Python?")
    print(f"\nAnswer: {answer}")
    print(f"Sources: {[doc.source for doc in sources]}")
```

---

## 2. Embeddings & Vector Representations

### Embedding Models

Embeddings convert text to dense vectors that capture semantic meaning:

```
Text: "The cat sat on the mat"
      ↓
Embedding: [0.234, -0.567, 0.891, ..., 0.123]  (384-dim vector)

Similarity: Text with similar meaning → Similar vectors
```

#### Example 3: Comparing Embedding Models

```python
# embedding_models_comparison.py
from sentence_transformers import SentenceTransformer
import numpy as np
from typing import Dict, List

class EmbeddingModelComparison:
    """Compare different embedding models"""
    
    MODELS = {
        "all-MiniLM-L6-v2": {
            "dimension": 384,
            "speed": "fast",
            "accuracy": "good",
            "size": "22MB",
            "use_case": "Fast local inference, development"
        },
        "all-mpnet-base-v2": {
            "dimension": 768,
            "speed": "medium",
            "accuracy": "very_good",
            "size": "438MB",
            "use_case": "Balanced quality/speed"
        },
        "multilingual-e5-base": {
            "dimension": 768,
            "speed": "medium",
            "accuracy": "very_good",
            "size": "460MB",
            "use_case": "Multilingual support (100+ languages)"
        },
        "bge-base-en-v1.5": {
            "dimension": 768,
            "speed": "medium",
            "accuracy": "excellent",
            "size": "438MB",
            "use_case": "Best accuracy for English"
        },
        "bge-large-en-v1.5": {
            "dimension": 1024,
            "speed": "slow",
            "accuracy": "excellent",
            "size": "1.3GB",
            "use_case": "Maximum accuracy (requires larger storage)"
        },
        "OpenAI ada-002": {
            "dimension": 1536,
            "speed": "fast",
            "accuracy": "excellent",
            "size": "API",
            "use_case": "Production use, no local setup"
        },
        "Cohere embed-english-v3.0": {
            "dimension": 1024,
            "speed": "fast",
            "accuracy": "excellent",
            "size": "API",
            "use_case": "Production use, enterprise support"
        }
    }
    
    @staticmethod
    def load_model(model_name: str) -> SentenceTransformer:
        """Load embedding model"""
        
        print(f"Loading {model_name}...")
        
        if model_name.startswith("OpenAI"):
            # Use OpenAI API
            from openai import OpenAI
            return OpenAIEmbeddings()
        
        elif model_name.startswith("Cohere"):
            # Use Cohere API
            from cohere import Cohere
            return CohereEmbeddings()
        
        else:
            # Load open-source model
            return SentenceTransformer(model_name)
    
    @staticmethod
    def compare_performance(texts: List[str]) -> Dict:
        """Compare performance across models"""
        
        results = {}
        
        for model_name, specs in EmbeddingModelComparison.MODELS.items():
            try:
                model = EmbeddingModelComparison.load_model(model_name)
                
                # Encode texts
                embeddings = model.encode(texts)
                
                # Calculate statistics
                results[model_name] = {
                    "specs": specs,
                    "embedding_shape": embeddings.shape,
                    "norm": np.linalg.norm(embeddings[0]),
                }
            
            except Exception as e:
                results[model_name] = {
                    "specs": specs,
                    "error": str(e)
                }
        
        return results
    
    @staticmethod
    def print_comparison():
        """Print comparison table"""
        
        print("\nEmbedding Models Comparison (2025-11-19)")
        print("=" * 100)
        print(f"{'Model':<35} {'Dim':<6} {'Speed':<10} {'Accuracy':<12} {'Size':<10}")
        print("-" * 100)
        
        for model_name, specs in EmbeddingModelComparison.MODELS.items():
            print(f"{model_name:<35} "
                  f"{specs['dimension']:<6} "
                  f"{specs['speed']:<10} "
                  f"{specs['accuracy']:<12} "
                  f"{specs['size']:<10}")
        
        print("\n" + "=" * 100)
        print("\nSelection Guide:")
        print("  Development:     all-MiniLM-L6-v2 (fast, small, good quality)")
        print("  Production:      bge-large-en-v1.5 or OpenAI ada-002 (best quality)")
        print("  Multilingual:    multilingual-e5-base (100+ languages)")
        print("  Balanced:        bge-base-en-v1.5 (quality/speed tradeoff)")


# Batch encoding for efficiency
class BatchEmbedding:
    """Efficient batch encoding"""
    
    def __init__(self, model_name: str = "all-MiniLM-L6-v2"):
        self.model = SentenceTransformer(model_name)
    
    def encode_batch(self, texts: List[str], batch_size: int = 32) -> np.ndarray:
        """Encode texts in batches for memory efficiency"""
        
        embeddings = self.model.encode(
            texts,
            batch_size=batch_size,
            show_progress_bar=True,
            convert_to_numpy=True
        )
        
        return embeddings
    
    def encode_streaming(self, texts: List[str], batch_size: int = 32):
        """Stream embeddings for large datasets"""
        
        for i in range(0, len(texts), batch_size):
            batch = texts[i:i + batch_size]
            embeddings = self.model.encode(batch, convert_to_numpy=True)
            yield embeddings


# Similarity calculation
class SimilaritySearch:
    """Calculate semantic similarity"""
    
    @staticmethod
    def cosine_similarity(vec1: np.ndarray, vec2: np.ndarray) -> float:
        """Cosine similarity between vectors"""
        
        dot_product = np.dot(vec1, vec2)
        magnitude1 = np.linalg.norm(vec1)
        magnitude2 = np.linalg.norm(vec2)
        
        return dot_product / (magnitude1 * magnitude2) if magnitude1 * magnitude2 > 0 else 0
    
    @staticmethod
    def find_most_similar(query_embedding: np.ndarray, 
                         corpus_embeddings: np.ndarray,
                         top_k: int = 5) -> List[Tuple[int, float]]:
        """Find most similar vectors"""
        
        # Calculate similarities
        similarities = []
        
        for i, corpus_embedding in enumerate(corpus_embeddings):
            sim = SimilaritySearch.cosine_similarity(query_embedding, corpus_embedding)
            similarities.append((i, sim))
        
        # Sort by similarity
        similarities.sort(key=lambda x: x[1], reverse=True)
        
        return similarities[:top_k]


if __name__ == "__main__":
    # Print comparison
    EmbeddingModelComparison.print_comparison()
```

### Using Embeddings in RAG

```python
# rag_with_embeddings.py
from sentence_transformers import SentenceTransformer

class EmbeddingRAG:
    """RAG system with proper embedding handling"""
    
    def __init__(self, embedding_model_name: str = "all-MiniLM-L6-v2"):
        self.embedding_model = SentenceTransformer(embedding_model_name)
        self.document_store = []
        self.embeddings_store = []
    
    def add_documents(self, documents: List[str]):
        """Encode and store documents"""
        
        print(f"Encoding {len(documents)} documents...")
        
        # Batch encode for efficiency
        embeddings = self.embedding_model.encode(
            documents,
            batch_size=32,
            show_progress_bar=True,
            convert_to_numpy=True
        )
        
        self.document_store.extend(documents)
        self.embeddings_store.extend(embeddings)
        
        print(f"Stored {len(documents)} documents with embeddings")
    
    def search(self, query: str, top_k: int = 5) -> List[Tuple[str, float]]:
        """Search for similar documents"""
        
        # Encode query
        query_embedding = self.embedding_model.encode(query)
        
        # Calculate similarities
        similarities = []
        
        for doc, doc_embedding in zip(self.document_store, self.embeddings_store):
            # Cosine similarity
            similarity = np.dot(query_embedding, doc_embedding) / (
                np.linalg.norm(query_embedding) * np.linalg.norm(doc_embedding) + 1e-8
            )
            similarities.append((doc, similarity))
        
        # Sort and return top-k
        similarities.sort(key=lambda x: x[1], reverse=True)
        
        return similarities[:top_k]
```

---

## 3. Vector Databases

### Database Comparison

| Database | Type | Pricing | Scale | Features |
|----------|------|---------|-------|----------|
| **Pinecone** | Cloud | $2.99-19.99/month | Billions | Hybrid search, metadata filtering |
| **Weaviate** | Open-source | Self-hosted | Millions | HNSW, BM25, GraphQL API |
| **Milvus** | Open-source | Self-hosted | Billions | Distributed, high-performance |
| **Chroma** | Embedded | Free | Millions | Simple, local, fast |
| **FAISS** | Embedded | Free | Billions | CPU/GPU, index compression |
| **Qdrant** | Cloud/Open | Free-$299/month | Billions | High-performance, filtering |

#### Example 4: Pinecone Integration

```python
# pinecone_rag.py
from pinecone import Pinecone, ServerlessSpec
from sentence_transformers import SentenceTransformer

class PineconeRAG:
    """RAG with Pinecone cloud vector database"""
    
    def __init__(self, api_key: str, index_name: str = "rag-index"):
        # Initialize Pinecone
        self.pc = Pinecone(api_key=api_key)
        
        # Create index if not exists
        self.create_index_if_needed(index_name)
        self.index = self.pc.Index(index_name)
        
        # Load embedding model
        self.embedding_model = SentenceTransformer("all-MiniLM-L6-v2")
    
    def create_index_if_needed(self, index_name: str):
        """Create Pinecone index"""
        
        existing_indexes = self.pc.list_indexes()
        
        if index_name not in existing_indexes:
            self.pc.create_index(
                name=index_name,
                dimension=384,  # all-MiniLM-L6-v2 dimension
                metric="cosine",
                spec=ServerlessSpec(
                    cloud="aws",
                    region="us-east-1"
                )
            )
            print(f"Created index: {index_name}")
    
    def index_documents(self, documents: List[str], batch_size: int = 100):
        """Index documents in Pinecone"""
        
        print(f"Indexing {len(documents)} documents to Pinecone...")
        
        # Encode in batches
        for i in range(0, len(documents), batch_size):
            batch = documents[i:i + batch_size]
            
            # Create embeddings
            embeddings = self.embedding_model.encode(batch)
            
            # Prepare vectors with metadata
            vectors_to_upsert = []
            
            for j, (doc, embedding) in enumerate(zip(batch, embeddings)):
                vectors_to_upsert.append({
                    "id": f"doc_{i + j}",
                    "values": embedding.tolist(),
                    "metadata": {
                        "content": doc[:1000],  # Store first 1000 chars
                        "index": i + j
                    }
                })
            
            # Upsert to Pinecone
            self.index.upsert(vectors=vectors_to_upsert)
        
        print(f"Indexed {len(documents)} documents")
    
    def search(self, query: str, top_k: int = 5) -> List[str]:
        """Search Pinecone index"""
        
        # Encode query
        query_embedding = self.embedding_model.encode(query).tolist()
        
        # Query index
        results = self.index.query(
            vector=query_embedding,
            top_k=top_k,
            include_metadata=True
        )
        
        # Extract documents
        documents = [
            match["metadata"]["content"]
            for match in results["matches"]
        ]
        
        return documents


# Statistics about Pinecone
print("""
Pinecone Features (November 2025):
- Cloud-hosted (no infrastructure management)
- Supports 4 billion+ vectors
- Hybrid search: Vector + keyword combined
- Metadata filtering: Filter results by attributes
- Low latency: < 100ms queries
- Scalability: Auto-scaling based on load
- Pricing: Pay per request, no monthly minimums

Best For:
- Production RAG systems
- Enterprise applications
- When you don't want to manage infrastructure
- Large-scale knowledge bases

Integration with LangChain:
from langchain.vectorstores import Pinecone
vector_store = Pinecone.from_documents(docs, embeddings, index_name="index")
""")
```

#### Example 5: Weaviate Integration

```python
# weaviate_rag.py
import weaviate
from weaviate.auth import Auth

class WeaviateRAG:
    """RAG with Weaviate open-source vector database"""
    
    def __init__(self, url: str = "http://localhost:8080", api_key: str = None):
        # Connect to Weaviate
        auth_config = Auth.api_key(api_key) if api_key else None
        
        self.client = weaviate.connect_to_local(auth_credentials=auth_config)
        self.collection_name = "Documents"
    
    def create_collection(self):
        """Create Weaviate collection"""
        
        from weaviate.collections.classes.config import Configure
        
        collections = self.client.collections.list_all()
        collection_names = [col.name for col in collections]
        
        if self.collection_name not in collection_names:
            self.client.collections.create(
                name=self.collection_name,
                vectorizer_config=Configure.Vectorizer.text2vec_transformers(),
                generative_config=Configure.Generative.openai(),
            )
            print(f"Created collection: {self.collection_name}")
    
    def add_documents(self, documents: List[str]):
        """Add documents to Weaviate"""
        
        collection = self.client.collections.get(self.collection_name)
        
        # Add documents (vectorization happens automatically)
        uuids = collection.data.insert_many([
            {
                "content": doc,
            }
            for doc in documents
        ])
        
        print(f"Added {len(uuids)} documents")
    
    def search(self, query: str, top_k: int = 5) -> List[str]:
        """Search documents using hybrid search"""
        
        collection = self.client.collections.get(self.collection_name)
        
        # Hybrid search (vector + BM25)
        response = collection.query.hybrid(
            query=query,
            limit=top_k,
            where=None,
            fusion_type="relative_score"  # Combine vector and BM25
        )
        
        documents = [
            obj.properties["content"]
            for obj in response.objects
        ]
        
        return documents
    
    def generate_answer(self, query: str, top_k: int = 5) -> str:
        """Generate answer using retrieved documents + LLM"""
        
        collection = self.client.collections.get(self.collection_name)
        
        # Hybrid search with generation
        response = collection.generate.hybrid(
            query=query,
            limit=top_k,
            single_prompt="Explain this: {content}"
        )
        
        return response.generated


# Weaviate characteristics
print("""
Weaviate Features (November 2025):
- Open-source (can self-host or use cloud)
- Hybrid search: Vector + keyword (BM25)
- GraphQL API: Flexible queries
- HNSW indexing: Fast approximate nearest neighbors
- Generative search: Generate answers using LLM
- Multi-shard: Distributed architecture
- Sub-second latency: Optimized for speed

Best For:
- Organizations wanting open-source control
- Complex queries with metadata filtering
- Hybrid search (combining semantic + keyword)
- When you want to self-host

Integration with LangChain:
from langchain.vectorstores import Weaviate
vector_store = Weaviate.from_documents(docs, embeddings, client=client)
""")
```

#### Example 6: Milvus for Large Scale

```python
# milvus_rag.py
from pymilvus import Collection, connections, FieldSchema, CollectionSchema, DataType

class MilvusRAG:
    """RAG with Milvus for billions of vectors"""
    
    def __init__(self, uri: str = "http://localhost:19530"):
        # Connect to Milvus
        connections.connect("default", uri=uri)
        
        self.collection_name = "documents"
        self.dimension = 384
    
    def create_collection(self):
        """Create Milvus collection with index"""
        
        # Define schema
        fields = [
            FieldSchema(name="id", dtype=DataType.INT64, is_primary=True, auto_id=True),
            FieldSchema(name="embedding", dtype=DataType.FLOAT_VECTOR, dim=self.dimension),
            FieldSchema(name="content", dtype=DataType.VARCHAR, max_length=5000),
            FieldSchema(name="source", dtype=DataType.VARCHAR, max_length=255),
        ]
        
        schema = CollectionSchema(fields=fields, description="Document embeddings")
        
        # Create collection
        collection = Collection(
            name=self.collection_name,
            schema=schema
        )
        
        # Create index
        index_params = {
            "index_type": "IvfFlat",  # Or "HNSW" for faster search
            "metric_type": "L2",  # Euclidean distance
            "params": {"nlist": 100}  # Number of clusters
        }
        
        collection.create_index(field_name="embedding", index_params=index_params)
        
        print(f"Created Milvus collection: {self.collection_name}")
    
    def insert_embeddings(self, embeddings: list, contents: list, sources: list):
        """Insert embeddings into Milvus"""
        
        collection = Collection(self.collection_name)
        
        # Insert data
        data = [embeddings, contents, sources]
        
        result = collection.insert(data)
        collection.flush()
        
        print(f"Inserted {len(embeddings)} vectors")
    
    def search(self, query_embedding: list, top_k: int = 5) -> List[dict]:
        """Search for similar documents"""
        
        collection = Collection(self.collection_name)
        collection.load()
        
        # Search
        results = collection.search(
            data=[query_embedding],
            anns_field="embedding",
            param={"metric_type": "L2", "params": {"nprobe": 10}},
            limit=top_k,
            output_fields=["content", "source"]
        )
        
        collection.release()
        
        return results[0]


# Milvus characteristics
print("""
Milvus Features (November 2025):
- Open-source, distributed vector database
- Handles billions of vectors
- HNSW and IvfFlat indexes
- Multiple distance metrics (L2, IP, cosine)
- Horizontal scaling across nodes
- Production-grade performance
- Active community development

Best For:
- Enterprise large-scale RAG
- Billions of documents
- Distributed deployments
- Self-hosted requirements
- Cost-effective at scale

Indexing Options:
- HNSW: Faster search, higher memory
- IvfFlat: Good balance of speed and memory
- GPU acceleration available
""")
```

---

## 4. Chunking & Data Preparation

### Document Chunking Strategies

How you split documents significantly affects RAG quality:

#### Example 7: Chunking Strategies

```python
# chunking_strategies.py
from typing import List
from abc import ABC, abstractmethod

class ChunkingStrategy(ABC):
    """Base class for chunking strategies"""
    
    @abstractmethod
    def chunk(self, text: str) -> List[str]:
        pass


class FixedSizeChunking(ChunkingStrategy):
    """Fixed-size chunks (simplest approach)"""
    
    def __init__(self, chunk_size: int = 512, overlap: int = 128):
        self.chunk_size = chunk_size
        self.overlap = overlap
    
    def chunk(self, text: str) -> List[str]:
        """Split text into fixed-size chunks"""
        
        chunks = []
        
        # Split by words
        words = text.split()
        
        for i in range(0, len(words), self.chunk_size - self.overlap):
            chunk = " ".join(words[i:i + self.chunk_size])
            if len(chunk) > 0:
                chunks.append(chunk)
        
        return chunks


class SemanticChunking(ChunkingStrategy):
    """Chunk by semantic boundaries (sentences/paragraphs)"""
    
    def __init__(self, chunk_size: int = 3, overlap: int = 1):
        """chunk_size = number of sentences per chunk"""
        self.chunk_size = chunk_size
        self.overlap = overlap
    
    def chunk(self, text: str) -> List[str]:
        """Split by sentence boundaries"""
        
        import re
        
        # Split by sentences
        sentences = re.split(r'(?<=[.!?])\s+', text)
        
        chunks = []
        
        for i in range(0, len(sentences), self.chunk_size - self.overlap):
            chunk_sentences = sentences[i:i + self.chunk_size]
            chunk = " ".join(chunk_sentences)
            if len(chunk) > 0:
                chunks.append(chunk)
        
        return chunks


class RecursiveChunking(ChunkingStrategy):
    """Recursively chunk by separators (most effective)"""
    
    def __init__(self, chunk_size: int = 512, overlap: int = 128):
        self.chunk_size = chunk_size
        self.overlap = overlap
        
        # Try these separators in order
        self.separators = [
            "\n\n",      # Paragraph
            "\n",        # Line
            ". ",        # Sentence
            " ",         # Word
            ""           # Character
        ]
    
    def chunk(self, text: str) -> List[str]:
        """Recursively chunk by separators"""
        
        good_splits = []
        separator = self.separators[-1]
        
        for _s in self.separators:
            if _s == "":
                separator = _s
                break
            
            if _s in text:
                separator = _s
                break
        
        # Split by separator
        if separator:
            splits = text.split(separator)
        else:
            splits = list(text)
        
        # Merge splits to chunk_size
        good_splits = []
        current_chunk = ""
        
        for split in splits:
            # Skip empty splits
            if not split.strip():
                continue
            
            # Check if adding this split exceeds chunk_size
            test_chunk = current_chunk + separator + split if current_chunk else split
            
            if len(test_chunk) < self.chunk_size:
                current_chunk = test_chunk
            else:
                if current_chunk:
                    good_splits.append(current_chunk)
                current_chunk = split
        
        if current_chunk:
            good_splits.append(current_chunk)
        
        return good_splits


class MetadataChunking:
    """Preserve document structure in chunks"""
    
    def __init__(self, chunk_size: int = 512):
        self.chunk_size = chunk_size
    
    def chunk_with_metadata(self, text: str, source: str, doc_type: str) -> List[dict]:
        """Chunk while preserving metadata"""
        
        chunker = RecursiveChunking(chunk_size=self.chunk_size)
        chunks = chunker.chunk(text)
        
        return [
            {
                "content": chunk,
                "source": source,
                "type": doc_type,
                "chunk_index": i,
                "total_chunks": len(chunks)
            }
            for i, chunk in enumerate(chunks)
        ]


class OverlapOptimization:
    """Determine optimal overlap for your use case"""
    
    @staticmethod
    def analyze_overlap_impact(text: str, chunk_size: int = 512) -> dict:
        """Test different overlap percentages"""
        
        overlaps = [0, 0.1, 0.2, 0.3, 0.5]
        results = {}
        
        for overlap_pct in overlaps:
            overlap = int(chunk_size * overlap_pct)
            
            chunker = FixedSizeChunking(chunk_size=chunk_size, overlap=overlap)
            chunks = chunker.chunk(text)
            
            results[f"{overlap_pct*100}%"] = {
                "num_chunks": len(chunks),
                "total_tokens": sum(len(c.split()) for c in chunks),
                "overlap_cost": (len(chunks) - 1) * overlap if len(chunks) > 1 else 0
            }
        
        return results


# Comparison of strategies
print("""
Chunking Strategy Comparison:

Fixed-Size:
+ Simple to implement
- Ignores document structure
- May split mid-sentence
→ Use: Simple documents, when speed matters

Semantic (Sentence-based):
+ Preserves semantic units
+ More interpretable chunks
- Chunk size varies
→ Use: Articles, documentation, when quality matters

Recursive:
+ Respects document structure
+ Automatically finds good split points
+ Optimal for most use cases
→ Use: Most RAG systems (RECOMMENDED)

Metadata-Aware:
+ Preserves source information
+ Enables metadata filtering
+ Better for complex documents
→ Use: Multi-source knowledge bases
""")
```

---

## 5. Retrieval & Ranking Strategies

### Hybrid Retrieval

Combine dense (vector) and sparse (BM25) retrieval:

#### Example 8: Hybrid Search

```python
# hybrid_retrieval.py
from typing import List, Tuple
import numpy as np

class HybridRetriever:
    """Combine vector + BM25 retrieval"""
    
    def __init__(self, vector_db, bm25_index):
        self.vector_db = vector_db
        self.bm25_index = bm25_index
    
    def retrieve(self, query: str, top_k: int = 10) -> List[Tuple[str, float]]:
        """Hybrid search: vector + BM25"""
        
        # Vector search
        vector_results = self.vector_db.search(query, top_k=top_k)
        vector_scores = {doc: score for doc, score in vector_results}
        
        # BM25 search
        bm25_results = self.bm25_index.search(query, top_k=top_k)
        bm25_scores = {doc: score for doc, score in bm25_results}
        
        # Combine scores
        all_docs = set(vector_scores.keys()) | set(bm25_scores.keys())
        
        combined_scores = {}
        
        for doc in all_docs:
            # Normalize scores to 0-1 range
            v_score = vector_scores.get(doc, 0) / (max(vector_scores.values()) + 1e-8)
            b_score = bm25_scores.get(doc, 0) / (max(bm25_scores.values()) + 1e-8)
            
            # Weighted combination (adjust weights for your use case)
            combined_scores[doc] = 0.6 * v_score + 0.4 * b_score
        
        # Sort by combined score
        ranked_results = sorted(
            combined_scores.items(),
            key=lambda x: x[1],
            reverse=True
        )[:top_k]
        
        return ranked_results


class BM25Retriever:
    """BM25 sparse retrieval baseline"""
    
    def __init__(self):
        from rank_bm25 import BM25Okapi
        self.bm25 = None
        self.corpus = []
    
    def index(self, documents: List[str]):
        """Index documents with BM25"""
        
        from rank_bm25 import BM25Okapi
        
        # Tokenize
        tokenized_corpus = [doc.lower().split() for doc in documents]
        
        # Create BM25 index
        self.bm25 = BM25Okapi(tokenized_corpus)
        self.corpus = documents
    
    def search(self, query: str, top_k: int = 5) -> List[Tuple[str, float]]:
        """BM25 search"""
        
        # Tokenize query
        tokenized_query = query.lower().split()
        
        # Get BM25 scores
        scores = self.bm25.get_scores(tokenized_query)
        
        # Get top-k
        top_indices = np.argsort(scores)[::-1][:top_k]
        
        results = [
            (self.corpus[i], scores[i])
            for i in top_indices
        ]
        
        return results


# When to use Hybrid Search
print("""
Hybrid Retrieval Advantages:

Vector Search:
+ Semantic matching ("cat" finds "feline")
+ Handles synonyms well
- Struggles with exact keyword matching
- More compute intensive

BM25 Search:
+ Exact keyword matching
+ Fast on structured queries
+ Great for technical docs
- Misses semantic relationships
- Fails on synonyms

Hybrid:
+ Combines both strengths
+ Better recall (fewer missed documents)
+ Better precision (fewer irrelevant results)
+ Recommended for production RAG

Weight Tuning (adjust for your domain):
- Technical docs: 60% BM25, 40% vector
- Creative content: 60% vector, 40% BM25
- Mixed: 50/50
""")
```

### Re-ranking with Cross-Encoders

Improve ranking quality with more powerful models:

#### Example 9: Cross-Encoder Re-ranking

```python
# crossencoder_reranking.py
from sentence_transformers import CrossEncoder
from typing import List, Tuple

class ReRanker:
    """Re-rank retrieved documents with cross-encoder"""
    
    def __init__(self, model_name: str = "cross-encoder/ms-marco-MiniLM-L-12-v2"):
        """Load cross-encoder model"""
        
        self.model = CrossEncoder(model_name)
    
    def rerank(self, query: str, documents: List[str], top_k: int = 5) -> List[Tuple[str, float]]:
        """Re-rank documents by relevance to query"""
        
        # Create pairs: (query, document)
        pairs = [[query, doc] for doc in documents]
        
        # Get relevance scores
        scores = self.model.predict(pairs)
        
        # Sort by score
        ranked_results = sorted(
            zip(documents, scores),
            key=lambda x: x[1],
            reverse=True
        )
        
        return ranked_results[:top_k]


class FullRAGWithReranking:
    """Complete RAG pipeline with re-ranking"""
    
    def __init__(self, retriever, llm):
        self.retriever = retriever
        self.llm = llm
        self.reranker = ReRanker()
    
    def answer(self, query: str) -> str:
        """RAG with re-ranking"""
        
        # Step 1: Initial retrieval (get more than needed)
        initial_results = self.retriever.retrieve(query, top_k=20)
        documents = [doc for doc, score in initial_results]
        
        # Step 2: Re-rank
        reranked = self.reranker.rerank(query, documents, top_k=5)
        top_documents = [doc for doc, score in reranked]
        
        # Step 3: Generate answer
        context = "\n".join(top_documents)
        prompt = f"Context:\n{context}\n\nQuestion: {query}\n\nAnswer:"
        
        answer = self.llm.generate(prompt)
        
        return answer


# Cross-Encoder Models
print("""
Best Cross-Encoder Models (2025):

Fast (For speed):
- cross-encoder/ms-marco-MiniLM-L-6-v2 (33M params, fast)

Balanced:
- cross-encoder/ms-marco-MiniLM-L-12-v2 (33M params)
- cross-encoder/mmarco-mMiniLMv2-L12-H384-v1 (multilingual)

High Quality:
- cross-encoder/ms-marco-MiniLM-L-12-v2-large (139M params)
- cross-encoder/qnli-distilroberta-base

Re-ranking Effect:
- Retrieval-only: ~70% precision@5
- + Re-ranking: ~85% precision@5
- Improvement: +15% in result quality
""")
```

---

## 6. Prompt Engineering for RAG

### Prompt Design

```python
# rag_prompt_engineering.py
from typing import List

class RAGPrompts:
    """Well-designed prompts for RAG"""
    
    @staticmethod
    def qa_prompt(context: str, question: str) -> str:
        """Standard QA prompt"""
        
        return f"""You are a helpful assistant. Answer the question based on the provided context. If you don't know the answer, say so.

Context:
{context}

Question: {question}

Answer:"""
    
    @staticmethod
    def qa_with_sources(context: str, question: str) -> str:
        """QA with source attribution"""
        
        return f"""Answer the question based on the context. Include source references.

Context:
{context}

Question: {question}

Answer with sources:"""
    
    @staticmethod
    def cot_prompt(context: str, question: str) -> str:
        """Chain-of-thought reasoning"""
        
        return f"""Answer step by step, referring to the context.

Context:
{context}

Question: {question}

Let me think through this step by step:
1. First, I'll identify the relevant information...
2. Then, I'll reason about...
3. Therefore...

Answer:"""
    
    @staticmethod
    def minimal_prompt(context: str, question: str) -> str:
        """Minimal, efficient prompt"""
        
        return f"""Context: {context}

Q: {question}
A:"""


# Example 10: Complete RAG prompt engineering
class RAGPromptOptimizer:
    """Optimize prompts for RAG"""
    
    @staticmethod
    def create_prompt(
        context: str,
        question: str,
        style: str = "standard",
        include_instructions: bool = True,
        language: str = "english"
    ) -> str:
        """Create optimized prompt"""
        
        system_prompt = {
            "standard": "You are a helpful assistant.",
            "expert": "You are an expert assistant with deep knowledge.",
            "concise": "Be concise and direct.",
            "detailed": "Provide detailed explanations."
        }
        
        instruction = ""
        
        if include_instructions:
            instruction = """
If the context doesn't contain the answer, say "I don't know based on the provided context."
Always prioritize information from the context over general knowledge.
Cite the sources when referencing the context."""
        
        return f"""{system_prompt.get(style, system_prompt['standard'])}

{instruction}

Context:
{context}

Question: {question}

Answer:"""


# Few-shot prompting
class FewShotRAG:
    """Few-shot examples for better RAG"""
    
    EXAMPLES = [
        {
            "context": "Python is a high-level programming language.",
            "question": "What is Python?",
            "answer": "Python is a high-level programming language."
        },
        {
            "context": "The capital of France is Paris. Paris is located on the Seine river.",
            "question": "Where is the capital of France?",
            "answer": "The capital of France is Paris, located on the Seine river."
        }
    ]
    
    @staticmethod
    def create_few_shot_prompt(context: str, question: str, num_examples: int = 2) -> str:
        """Create prompt with few-shot examples"""
        
        examples_text = ""
        
        for i, example in enumerate(FewShotRAG.EXAMPLES[:num_examples]):
            examples_text += f"""
Example {i+1}:
Context: {example['context']}
Question: {example['question']}
Answer: {example['answer']}
"""
        
        return f"""Answer the question based on the provided context.

{examples_text}

Now, answer this:
Context: {context}
Question: {question}
Answer:"""
```

---

## 7. Advanced Patterns

### Multi-hop Retrieval

Handle complex questions requiring multiple retrieval steps:

#### Example 11: Multi-hop RAG

```python
# multihop_rag.py
from typing import List

class MultiHopRetriever:
    """Retrieve for complex, multi-step questions"""
    
    def __init__(self, retriever, llm):
        self.retriever = retriever
        self.llm = llm
    
    def decompose_question(self, question: str) -> List[str]:
        """Break complex question into sub-questions"""
        
        prompt = f"""Break this question into simpler sub-questions:

Question: {question}

Sub-questions:
1. """
        
        response = self.llm.generate(prompt)
        
        # Parse sub-questions
        sub_questions = response.split('\n')
        sub_questions = [q.strip() for q in sub_questions if q.strip()]
        
        return sub_questions[:5]  # Limit to 5 hops
    
    def multi_hop_retrieval(self, question: str) -> List[str]:
        """Retrieve context for multi-hop question"""
        
        # Decompose
        sub_questions = self.decompose_question(question)
        
        all_documents = []
        
        # Retrieve for each sub-question
        for sub_q in sub_questions:
            docs = self.retriever.retrieve(sub_q, top_k=3)
            all_documents.extend([doc for doc, _ in docs])
        
        # Deduplicate
        all_documents = list(set(all_documents))
        
        return all_documents
    
    def answer_multihop_question(self, question: str) -> str:
        """Answer complex question"""
        
        # Multi-hop retrieval
        documents = self.multi_hop_retrieval(question)
        context = "\n".join(documents)
        
        # Generate answer
        prompt = f"""Answer based on this context:

Context:
{context}

Question: {question}

Answer:"""
        
        answer = self.llm.generate(prompt)
        
        return answer
```

### Self-Refining RAG

Improve answers through iterative refinement:

```python
# self_refining_rag.py
class SelfRefiningRAG:
    """Refine answers by checking for hallucinations"""
    
    def __init__(self, retriever, llm):
        self.retriever = retriever
        self.llm = llm
    
    def answer_with_self_refinement(self, question: str, max_iterations: int = 3) -> str:
        """Iteratively refine answer"""
        
        # Initial answer
        initial_docs = self.retriever.retrieve(question, top_k=5)
        context = "\n".join([doc for doc, _ in initial_docs])
        
        answer = self.llm.generate(f"Question: {question}\nContext: {context}\nAnswer:")
        
        # Refine
        for i in range(max_iterations - 1):
            # Check for hallucinations
            check_prompt = f"""Check if this answer is grounded in the context:

Answer: {answer}

Context: {context}

Are there any claims not supported by context? List them:"""
            
            hallucinations = self.llm.generate(check_prompt)
            
            if "none" in hallucinations.lower() or "no" in hallucinations.lower():
                break  # Good answer, stop refining
            
            # Refine answer
            refine_prompt = f"""Refine this answer to remove unsupported claims:

Original answer: {answer}

Unsupported claims: {hallucinations}

Context: {context}

Refined answer:"""
            
            answer = self.llm.generate(refine_prompt)
        
        return answer
```

---

## 8. Evaluation & Testing

### RAG Metrics

#### Example 12: Evaluation Metrics

```python
# rag_evaluation.py
import numpy as np
from typing import List, Set

class RAGEvaluator:
    """Comprehensive RAG evaluation"""
    
    @staticmethod
    def hit_rate(retrieved_docs: List[str], relevant_docs: Set[str]) -> float:
        """Hit Rate: % of queries where relevant doc was retrieved"""
        
        retrieved_set = set(retrieved_docs)
        
        if not relevant_docs:
            return 0.0
        
        hits = len(retrieved_set & relevant_docs)
        
        return hits / len(relevant_docs)
    
    @staticmethod
    def mean_reciprocal_rank(retrieved_docs: List[str], relevant_doc: str) -> float:
        """MRR: Position of first relevant document"""
        
        for i, doc in enumerate(retrieved_docs):
            if doc == relevant_doc:
                return 1.0 / (i + 1)
        
        return 0.0
    
    @staticmethod
    def ndcg(retrieved_docs: List[str], relevant_docs: Set[str], k: int = 10) -> float:
        """Normalized DCG: Quality of ranking"""
        
        # DCG
        dcg = 0.0
        for i, doc in enumerate(retrieved_docs[:k]):
            relevance = 1.0 if doc in relevant_docs else 0.0
            dcg += relevance / np.log2(i + 2)
        
        # Ideal DCG
        idcg = sum(1.0 / np.log2(i + 2) for i in range(min(len(relevant_docs), k)))
        
        return dcg / idcg if idcg > 0 else 0.0
    
    @staticmethod
    def precision_at_k(retrieved_docs: List[str], relevant_docs: Set[str], k: int = 5) -> float:
        """Precision@K: % of top-K results that are relevant"""
        
        retrieved_set = set(retrieved_docs[:k])
        relevant_retrieved = len(retrieved_set & relevant_docs)
        
        return relevant_retrieved / k if k > 0 else 0.0
    
    @staticmethod
    def recall_at_k(retrieved_docs: List[str], relevant_docs: Set[str], k: int = 5) -> float:
        """Recall@K: % of relevant docs found in top-K"""
        
        retrieved_set = set(retrieved_docs[:k])
        relevant_retrieved = len(retrieved_set & relevant_docs)
        
        return relevant_retrieved / len(relevant_docs) if relevant_docs else 0.0
    
    @staticmethod
    def f1_at_k(retrieved_docs: List[str], relevant_docs: Set[str], k: int = 5) -> float:
        """F1@K: Harmonic mean of precision and recall"""
        
        precision = RAGEvaluator.precision_at_k(retrieved_docs, relevant_docs, k)
        recall = RAGEvaluator.recall_at_k(retrieved_docs, relevant_docs, k)
        
        if precision + recall == 0:
            return 0.0
        
        return 2 * (precision * recall) / (precision + recall)


class GenerationEvaluator:
    """Evaluate LLM-generated answers"""
    
    @staticmethod
    def exact_match(prediction: str, reference: str) -> float:
        """1.0 if exact match, 0.0 otherwise"""
        
        return 1.0 if prediction.lower() == reference.lower() else 0.0
    
    @staticmethod
    def bleu_score(prediction: str, reference: str) -> float:
        """BLEU score for similarity"""
        
        from nltk.translate.bleu_score import sentence_bleu
        
        pred_tokens = prediction.lower().split()
        ref_tokens = reference.lower().split()
        
        return sentence_bleu([ref_tokens], pred_tokens)
    
    @staticmethod
    def rouge_score(prediction: str, reference: str) -> dict:
        """ROUGE score (recall-oriented)"""
        
        from rouge_score import rouge_scorer
        
        scorer = rouge_scorer.RougeScorer(['rouge1', 'rouge2', 'rougeL'])
        scores = scorer.score(reference, prediction)
        
        return {k: v.fmeasure for k, v in scores.items()}
    
    @staticmethod
    def bert_score(prediction: str, reference: str) -> float:
        """BERTScore for semantic similarity"""
        
        from bert_score import score
        
        P, R, F1 = score([prediction], [reference], lang='en', verbose=False)
        
        return F1.item()


# Evaluation benchmark
print("""
RAG Evaluation Metrics:

Retrieval Quality:
- Hit Rate: % of queries with relevant doc retrieved (target: > 90%)
- MRR@10: Position of first relevant doc (target: > 0.7)
- NDCG@10: Quality of ranking (target: > 0.6)
- Precision@5: % of top-5 results relevant (target: > 70%)
- Recall@5: % of relevant docs found in top-5 (target: > 60%)

Generation Quality:
- BLEU: N-gram overlap with reference (target: > 0.3)
- ROUGE: Recall-oriented overlap (target: > 0.4)
- BERTScore: Semantic similarity (target: > 0.8)

Benchmarks:
- Good RAG system: Hit Rate > 85%, NDCG > 0.6, BLEU > 0.25
- Excellent system: Hit Rate > 95%, NDCG > 0.75, BERTScore > 0.85
""")
```

---

## 9. Production Deployment

### Architecture Overview

```python
# production_rag_architecture.py
class ProductionRAG:
    """Enterprise-grade RAG architecture"""
    
    def __init__(self):
        """Initialize all components"""
        
        # Caching
        self.embedding_cache = {}  # Redis in production
        self.response_cache = {}   # Redis in production
        
        # Components
        self.embedding_model = None  # Load once
        self.vector_db = None        # Cloud or self-hosted
        self.llm = None              # API or local
        self.reranker = None         # Optional, for quality
    
    def retrieve_with_cache(self, query: str, ttl: int = 3600) -> List[str]:
        """Retrieve with caching"""
        
        # Check response cache
        cache_key = f"response_{query}"
        
        if cache_key in self.response_cache:
            return self.response_cache[cache_key]
        
        # Check embedding cache
        embedding_cache_key = f"embedding_{query}"
        
        if embedding_cache_key in self.embedding_cache:
            embedding = self.embedding_cache[embedding_cache_key]
        else:
            embedding = self.embedding_model.encode(query)
            self.embedding_cache[embedding_cache_key] = embedding
        
        # Search vector DB
        results = self.vector_db.search(embedding, top_k=5)
        
        # Cache results
        self.response_cache[cache_key] = results
        
        return results


class RateLimiter:
    """Rate limiting for production"""
    
    def __init__(self, max_requests: int = 100, window: int = 60):
        self.max_requests = max_requests
        self.window = window
        self.requests = {}
    
    def is_allowed(self, user_id: str) -> bool:
        """Check if request is allowed"""
        
        import time
        current_time = time.time()
        
        if user_id not in self.requests:
            self.requests[user_id] = []
        
        # Remove old requests
        self.requests[user_id] = [
            t for t in self.requests[user_id]
            if current_time - t < self.window
        ]
        
        # Check limit
        if len(self.requests[user_id]) < self.max_requests:
            self.requests[user_id].append(current_time)
            return True
        
        return False


class MetricsCollector:
    """Track RAG metrics for monitoring"""
    
    def __init__(self):
        self.metrics = {
            "total_queries": 0,
            "total_latency": 0,
            "retrieval_quality": [],
            "generation_quality": [],
            "cache_hits": 0,
            "cache_misses": 0
        }
    
    def record_query(self, latency: float, retrieval_score: float, generation_score: float, cache_hit: bool):
        """Record metrics for one query"""
        
        self.metrics["total_queries"] += 1
        self.metrics["total_latency"] += latency
        self.metrics["retrieval_quality"].append(retrieval_score)
        self.metrics["generation_quality"].append(generation_score)
        
        if cache_hit:
            self.metrics["cache_hits"] += 1
        else:
            self.metrics["cache_misses"] += 1
    
    def get_report(self) -> dict:
        """Generate performance report"""
        
        avg_latency = (self.metrics["total_latency"] / self.metrics["total_queries"]
                      if self.metrics["total_queries"] > 0 else 0)
        
        cache_hit_rate = (self.metrics["cache_hits"] / 
                         (self.metrics["cache_hits"] + self.metrics["cache_misses"])
                         if self.metrics["cache_hits"] + self.metrics["cache_misses"] > 0 else 0)
        
        return {
            "total_queries": self.metrics["total_queries"],
            "avg_latency_ms": avg_latency * 1000,
            "cache_hit_rate": cache_hit_rate,
            "avg_retrieval_score": np.mean(self.metrics["retrieval_quality"]) if self.metrics["retrieval_quality"] else 0,
            "avg_generation_score": np.mean(self.metrics["generation_quality"]) if self.metrics["generation_quality"] else 0
        }
```

---

## Best Practices

### Data Quality

RAG quality depends entirely on your document quality:

```
Poor documents → Poor retrieval → Poor answers
```

**Checklist**:
- Remove duplicates (biases toward repeated info)
- Verify accuracy (fact-check before indexing)
- Clean formatting (consistent structure helps)
- Update regularly (knowledge grows stale)
- Check for private information (PII, credentials)

### Monitoring

```python
# monitoring.py
class RAGMonitor:
    """Monitor RAG system health"""
    
    @staticmethod
    def log_metrics(
        query: str,
        retrieved_docs: List[str],
        answer: str,
        latency_ms: float,
        user_id: str = "unknown"
    ):
        """Log query metrics"""
        
        import json
        from datetime import datetime
        
        log_entry = {
            "timestamp": datetime.now().isoformat(),
            "user_id": user_id,
            "query": query,
            "num_docs_retrieved": len(retrieved_docs),
            "answer_length": len(answer),
            "latency_ms": latency_ms
        }
        
        print(json.dumps(log_entry))
    
    @staticmethod
    def alert_on_issues(
        retrieval_quality: float,
        generation_quality: float,
        latency_ms: float
    ):
        """Alert on performance issues"""
        
        if retrieval_quality < 0.5:
            print(f"WARNING: Low retrieval quality ({retrieval_quality})")
        
        if generation_quality < 0.6:
            print(f"WARNING: Low generation quality ({generation_quality})")
        
        if latency_ms > 5000:
            print(f"WARNING: High latency ({latency_ms}ms)")
```

---

## TRUST 5 Compliance

### Test-First

RAG systems need rigorous testing:

```python
def test_rag_quality():
    """Test RAG system meets quality thresholds"""
    
    # Test 1: Retrieval accuracy
    test_queries = [...]
    for query, expected_docs in test_queries:
        retrieved = rag.retrieve(query)
        hit_rate = len(set(retrieved) & set(expected_docs)) / len(expected_docs)
        assert hit_rate > 0.8, f"Hit rate too low: {hit_rate}"
    
    # Test 2: Generation quality
    for query, expected_answer in test_queries:
        answer = rag.answer(query)
        similarity = calculate_similarity(answer, expected_answer)
        assert similarity > 0.7, f"Answer quality too low: {similarity}"
    
    # Test 3: Latency
    import time
    start = time.time()
    rag.answer("Test query")
    latency = time.time() - start
    assert latency < 5, f"Latency too high: {latency}s"
```

### Readable

Clear documentation and code structure:

```python
# Good: Self-documenting
rag = RAGSystem(
    embedding_model="all-MiniLM-L6-v2",  # Fast local model
    vector_db="pinecone",                 # Cloud, managed
    llm="gpt-4",                          # High quality generation
    cache=redis_client                    # Speed up retrievals
)

# Bad: Cryptic
rag = RAGSystem(
    em="minilm",
    db="pdb",
    llm="4",
    c=rc
)
```

### Unified

Consistent patterns across RAG implementations:

```
All RAG systems follow:
1. Load/prepare documents
2. Create embeddings
3. Store in vector database
4. Accept user query
5. Retrieve relevant docs
6. Rerank (optional)
7. Generate answer
8. Log metrics
```

### Secured

Security best practices:

```python
# Do:
- Sanitize user inputs
- Filter PII from documents
- Use HTTPS for API calls
- Audit document access
- Encrypt sensitive data

# Don't:
- Log full documents or API keys
- Expose internal system prompts
- Allow arbitrary document uploads
- Skip authentication
```

### Trackable

Version and track RAG improvements:

```
RAG-DEPLOYMENT-001: Initial RAG system
├─ Retrieval: Hit Rate 85%, NDCG 0.62
├─ Generation: BLEU 0.24, BERTScore 0.81
├─ Latency: 450ms
└─ Cost: $200/month

RAG-DEPLOYMENT-002: Added re-ranking
├─ Retrieval: Hit Rate 92%, NDCG 0.73 (+11%)
├─ Generation: BLEU 0.31, BERTScore 0.87 (+7%)
├─ Latency: 650ms (+44%)
└─ Cost: $220/month (+10%)

Decision: Deploy RAG-002 (quality improvement > latency cost)
```

---

## Quick Reference: Technology Stack (November 2025)

### Embedding Models
- **Local**: all-MiniLM-L6-v2 (fast), bge-base-en-v1.5 (quality), multilingual-e5-base (multilingual)
- **API**: OpenAI ada-002, Cohere embed-english-v3.0

### Vector Databases
- **Cloud**: Pinecone (managed, hybrid search)
- **Open-source**: Weaviate (GraphQL, hybrid), Milvus (large-scale), Chroma (local)

### Frameworks
- **LangChain**: Most popular, extensive integrations
- **LlamaIndex**: Document-focused, better for retrieval
- **Haystack**: Production-ready, good community

### Re-ranking
- **cross-encoder/ms-marco-MiniLM-L-12-v2**: Fast and effective
- **BGE Reranker**: Latest, best quality

### LLMs for Generation
- **GPT-4**: Highest quality but expensive
- **Claude 3**: Excellent reasoning, moderate cost
- **Llama 3.1**: Open-source, cost-effective

---

**Last Updated**: 2025-11-19
**Version**: 4.0.0
**Status**: Stable (Production-ready)
**Language**: English
**Framework**: LangChain 0.2+, LlamaIndex 0.10+, Pinecone 3.0+, Weaviate latest, sentence-transformers 3.0+
**Deployment**: AWS, GCP, Azure, self-hosted options
**Enterprise Ready**: Yes (monitoring, caching, rate limiting, security)

