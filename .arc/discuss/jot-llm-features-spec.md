# Jot LLM/Agent Integration Specification

## Overview

This document details how Jot makes documentation maximally useful for LLMs, agents, and AI systems through structured data, semantic markup, and purpose-built APIs.

## Core LLM-Friendly Features

### 1. Structured Data Export

#### JSON Export Format
```json
{
  "version": "1.0",
  "generated": "2024-01-13T12:00:00Z",
  "project": {
    "name": "Project Name",
    "description": "Project description",
    "metadata": {}
  },
  "documents": [
    {
      "id": "unique-doc-id",
      "path": "docs/getting-started.md",
      "title": "Getting Started",
      "content": "Raw markdown content",
      "html": "Rendered HTML content",
      "sections": [
        {
          "id": "section-id",
          "title": "Installation",
          "level": 2,
          "content": "Section content",
          "start_line": 10,
          "end_line": 25
        }
      ],
      "metadata": {
        "author": "John Doe",
        "tags": ["tutorial", "beginner"],
        "last_modified": "2024-01-13T11:00:00Z"
      },
      "links": {
        "internal": ["./installation.md", "./configuration.md"],
        "external": ["https://example.com"],
        "anchors": ["#prerequisites", "#next-steps"]
      },
      "code_blocks": [
        {
          "language": "bash",
          "content": "npm install jot",
          "line_number": 15
        }
      ]
    }
  ],
  "index": {
    "keywords": ["keyword1", "keyword2"],
    "concepts": ["concept1", "concept2"],
    "relationships": [
      {
        "source": "doc-id-1",
        "target": "doc-id-2",
        "type": "references"
      }
    ]
  }
}
```

### 2. Semantic Markup

#### Microdata Integration
```html
<article itemscope itemtype="http://schema.org/TechArticle">
  <h1 itemprop="headline">Getting Started with Jot</h1>
  <div itemprop="author" itemscope itemtype="http://schema.org/Person">
    <span itemprop="name">John Doe</span>
  </div>
  <time itemprop="datePublished" datetime="2024-01-13">January 13, 2024</time>
  <div itemprop="articleBody">
    <!-- Content -->
  </div>
</article>
```

#### JSON-LD Support
```html
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  "name": "Jot",
  "applicationCategory": "DeveloperApplication",
  "operatingSystem": "Windows, macOS, Linux",
  "offers": {
    "@type": "Offer",
    "price": "0",
    "priceCurrency": "USD"
  }
}
</script>
```

### 3. Vector Embedding Support

#### Embedding Generation
```go
type EmbeddingConfig struct {
    Model      string   // "sentence-transformers/all-MiniLM-L6-v2"
    Dimensions int      // 384
    ChunkSize  int      // 512 tokens
    Overlap    int      // 128 tokens
}

type DocumentEmbedding struct {
    DocumentID string
    Chunks     []ChunkEmbedding
}

type ChunkEmbedding struct {
    ID         string
    Text       string
    Vector     []float32
    Metadata   map[string]interface{}
}
```

#### Export Formats
- **Pinecone**: Direct upload format
- **Weaviate**: Schema-compliant JSON
- **ChromaDB**: Collection-ready format
- **FAISS**: Index-compatible vectors

### 4. Agent-Specific APIs

#### REST Endpoints
```
# Document retrieval with context
GET /api/agent/context/:doc_id
Response: {
  "document": {...},
  "related": [...],
  "breadcrumb": [...],
  "snippets": [...]
}

# Semantic search
POST /api/agent/search
Body: {
  "query": "how to install",
  "context": "user is beginner",
  "limit": 10
}

# Question answering
POST /api/agent/qa
Body: {
  "question": "What are the system requirements?",
  "context_docs": ["doc-1", "doc-2"]
}
```

#### GraphQL Schema
```graphql
type Query {
  document(id: ID!): Document
  documents(filter: DocumentFilter): [Document]
  search(query: String!, limit: Int): SearchResult
  related(id: ID!, limit: Int): [Document]
}

type Document {
  id: ID!
  title: String!
  content: String!
  sections: [Section]
  metadata: Metadata
  embeddings: [Embedding]
}

type Section {
  id: ID!
  title: String!
  content: String!
  level: Int!
  codeBlocks: [CodeBlock]
}
```

### 5. Context Window Optimization

#### Smart Chunking
```go
type ChunkingStrategy struct {
    Method        string // "semantic", "sliding", "recursive"
    MaxTokens     int    // 2048
    PreserveCode  bool   // Keep code blocks intact
    IncludeContext bool  // Add section headers
}

func (s *ChunkingStrategy) Chunk(doc Document) []Chunk {
    // Intelligent chunking that preserves meaning
}
```

#### Context Compression
```json
{
  "mode": "summary",
  "document": "full-doc-id",
  "summary": "Compressed document summary preserving key information",
  "key_points": ["point1", "point2"],
  "code_examples": ["essential code only"],
  "token_count": 500
}
```

### 6. LLM-Friendly Features

#### Prompt Templates
```yaml
templates:
  document_summary:
    system: "You are a documentation assistant"
    template: |
      Document: {title}
      Path: {path}
      Content: {content}
      
      Task: {task}
  
  code_explanation:
    system: "You are a code documentation expert"
    template: |
      Code Block:
      ```{language}
      {code}
      ```
      Context: {context}
      
      Explain this code:
```

#### Conversation Memory
```go
type ConversationContext struct {
    SessionID   string
    History     []Message
    CurrentDoc  string
    ViewedDocs  []string
    Preferences map[string]interface{}
}
```

### 7. Integration Patterns

#### LangChain Integration
```python
from langchain.document_loaders import JotLoader
from langchain.text_splitter import JotSplitter
from langchain.vectorstores import JotVectorStore

# Load documentation
loader = JotLoader("http://localhost:8080/api")
docs = loader.load()

# Process with LangChain
splitter = JotSplitter(chunk_size=1000)
chunks = splitter.split_documents(docs)

# Store in vector database
vectorstore = JotVectorStore.from_documents(chunks)
```

#### OpenAI Function Calling
```json
{
  "name": "search_documentation",
  "description": "Search project documentation",
  "parameters": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "Search query"
      },
      "filters": {
        "type": "object",
        "properties": {
          "tags": {"type": "array", "items": {"type": "string"}},
          "section": {"type": "string"}
        }
      }
    }
  }
}
```

### 8. Analytics & Feedback

#### Usage Tracking
```go
type AgentInteraction struct {
    AgentID    string
    Timestamp  time.Time
    Action     string // "search", "retrieve", "generate"
    DocumentID string
    Success    bool
    Feedback   *Feedback
}

type Feedback struct {
    Helpful    bool
    Accuracy   int // 1-5
    Missing    string
    Suggestion string
}
```

## Implementation Priority

1. **Phase 1**: Basic JSON/YAML export
2. **Phase 2**: REST API with search
3. **Phase 3**: Embedding generation
4. **Phase 4**: Advanced agent APIs
5. **Phase 5**: Analytics and optimization

## Performance Considerations

- **Caching**: Aggressive caching of embeddings and processed data
- **Streaming**: Support for streaming large responses
- **Pagination**: Efficient handling of large document sets
- **Rate Limiting**: Protect against abuse while allowing legitimate use

## Security & Privacy

- **API Keys**: Required for agent access
- **Rate Limiting**: Per-key limits
- **Data Filtering**: Exclude sensitive content
- **Audit Logging**: Track all agent interactions

## Success Metrics

- Agent query response time < 100ms
- Embedding generation < 1s per document
- 95%+ accuracy in semantic search
- Support for 10k+ concurrent agent requests
- < 5% token waste in responses