-- deployments/migrations/000001_create_schema.up.sql

-- 1. Create profiles table
CREATE TABLE profiles (
    id CHAR(26) PRIMARY KEY,
    full_name TEXT NOT NULL,
    headline TEXT,
    bio TEXT,
    email TEXT,
    phone TEXT,
    location TEXT,
    github TEXT,
    linkedin TEXT,
    website TEXT,
    avatar TEXT,
    resume_url TEXT,
    availability TEXT, -- Available, Busy, Not Looking
    timezone TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Create experiences table
CREATE TABLE experiences (
    id CHAR(26) PRIMARY KEY,
    company TEXT NOT NULL,
    position TEXT NOT NULL,
    employment_type TEXT,
    start_date DATE,
    end_date DATE,
    current_job BOOLEAN DEFAULT FALSE,
    location TEXT,
    description TEXT,
    display_order INT DEFAULT 0,
    company_logo TEXT,
    skills JSONB,
    remote_type TEXT, -- Remote, Hybrid, Onsite
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Create projects table
CREATE TABLE projects (
    id CHAR(26) PRIMARY KEY,
    slug TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    summary TEXT,
    description TEXT,
    architecture TEXT,
    repository_url TEXT,
    demo_url TEXT,
    thumbnail TEXT,
    featured BOOLEAN DEFAULT FALSE,
    status TEXT DEFAULT 'Draft', -- Draft, Published, Archived
    github_stars INT DEFAULT 0,
    github_last_commit TIMESTAMPTZ,
    read_time INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Create technologies table
CREATE TABLE technologies (
    id CHAR(26) PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    category TEXT,
    icon TEXT,
    color TEXT,
    official_url TEXT,
    logo TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. Create project_technologies table (Many-to-many)
CREATE TABLE project_technologies (
    project_id CHAR(26) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    technology_id CHAR(26) NOT NULL REFERENCES technologies(id) ON DELETE CASCADE,
    display_order INT DEFAULT 0,
    PRIMARY KEY (project_id, technology_id)
);

-- 6. Create certificates table
CREATE TABLE certificates (
    id CHAR(26) PRIMARY KEY,
    title TEXT NOT NULL,
    issuer TEXT NOT NULL,
    issue_date DATE,
    expiration_date DATE,
    credential_id TEXT,
    credential_url TEXT,
    thumbnail TEXT,
    skills JSONB,
    issuer_logo TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 7. Create skills table
CREATE TABLE skills (
    id CHAR(26) PRIMARY KEY,
    technology_id CHAR(26) NOT NULL REFERENCES technologies(id) ON DELETE CASCADE,
    display_order INT DEFAULT 0,
    level TEXT,
    years NUMERIC(4, 1) DEFAULT 0.0,
    favorite BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 8. Create knowledge_documents table
CREATE TABLE knowledge_documents (
    id CHAR(26) PRIMARY KEY,
    source_type TEXT NOT NULL, -- profile, experience, project, certificate, blog, manual
    source_id CHAR(26),
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    checksum TEXT NOT NULL,
    version INT DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'Pending', -- Pending, Embedding, Embedded, Failed
    embedding_model TEXT,
    last_embedded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 9. Create knowledge_chunks table
CREATE TABLE knowledge_chunks (
    id CHAR(26) PRIMARY KEY,
    document_id CHAR(26) NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    content TEXT NOT NULL,
    token_count INT,
    embedding_model TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 10. Create visitors table
CREATE TABLE visitors (
    id CHAR(26) PRIMARY KEY,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    total_messages INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 11. Create visitor_memories table (facts)
CREATE TABLE visitor_memories (
    id CHAR(26) PRIMARY KEY,
    visitor_id CHAR(26) NOT NULL REFERENCES visitors(id) ON DELETE CASCADE,
    category TEXT,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    confidence NUMERIC(5, 4),
    last_confirmed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (visitor_id, key)
);

-- 12. Create visitor_knowledge table (summary of visitor chat)
CREATE TABLE visitor_knowledge (
    id CHAR(26) PRIMARY KEY,
    visitor_id CHAR(26) NOT NULL REFERENCES visitors(id) ON DELETE CASCADE,
    category TEXT,
    memory_text TEXT NOT NULL,
    importance INT DEFAULT 3,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 13. Create ai_models table
CREATE TABLE ai_models (
    id CHAR(26) PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    provider TEXT NOT NULL,
    temperature NUMERIC(3, 2) DEFAULT 0.7,
    max_tokens INT,
    context_window INT,
    supports_tools BOOLEAN DEFAULT FALSE,
    supports_stream BOOLEAN DEFAULT FALSE,
    enabled BOOLEAN DEFAULT TRUE
);

-- 14. Create prompts table
CREATE TABLE prompts (
    id CHAR(26) PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    system_prompt TEXT NOT NULL,
    description TEXT,
    model_id CHAR(26) REFERENCES ai_models(id) ON DELETE SET NULL,
    active BOOLEAN DEFAULT FALSE,
    version INT DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 15. Create chat_sessions table
CREATE TABLE chat_sessions (
    id CHAR(26) PRIMARY KEY,
    visitor_id CHAR(26) NOT NULL REFERENCES visitors(id) ON DELETE CASCADE,
    prompt_id CHAR(26) REFERENCES prompts(id) ON DELETE SET NULL,
    title TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 16. Create chat_messages table
CREATE TABLE chat_messages (
    id CHAR(26) PRIMARY KEY,
    session_id CHAR(26) NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role TEXT NOT NULL, -- system, user, assistant, tool
    content TEXT NOT NULL,
    model TEXT,
    prompt_tokens INT,
    completion_tokens INT,
    latency_ms INT,
    status TEXT NOT NULL DEFAULT 'Pending', -- Pending, Streaming, Completed, Error
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 17. Create ai_tools table
CREATE TABLE ai_tools (
    id CHAR(26) PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    tool_type TEXT NOT NULL,
    config JSONB,
    description TEXT,
    enabled BOOLEAN DEFAULT TRUE
);

-- 18. Create outbox_events table
CREATE TABLE outbox_events (
    id CHAR(26) PRIMARY KEY,
    aggregate TEXT NOT NULL,
    aggregate_id CHAR(26) NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    published BOOLEAN DEFAULT FALSE,
    retry_count INT DEFAULT 0,
    failed_reason TEXT,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 19. Create embedding_profiles table
CREATE TABLE embedding_profiles (
    id CHAR(26) PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    dimension INT NOT NULL,
    metric_type TEXT NOT NULL DEFAULT 'COSINE',
    knowledge_collection TEXT NOT NULL,
    visitor_collection TEXT NOT NULL,
    enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 20. Create audit_logs table
CREATE TABLE audit_logs (
    id CHAR(26) PRIMARY KEY,
    table_name TEXT NOT NULL,
    record_id CHAR(26) NOT NULL,
    action TEXT NOT NULL,
    actor TEXT,
    before JSONB,
    after JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 21. Create retrieval_logs table
CREATE TABLE retrieval_logs (
    id CHAR(26) PRIMARY KEY,
    session_id CHAR(26) REFERENCES chat_sessions(id) ON DELETE SET NULL,
    query TEXT NOT NULL,
    rewritten_query TEXT,
    documents JSONB,
    chunks JSONB,
    model TEXT,
    response_time_ms INT,
    search_time_ms INT,
    llm_time_ms INT,
    total_time_ms INT,
    top_k INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
