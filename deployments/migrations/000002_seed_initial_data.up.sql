-- deployments/migrations/000002_seed_initial_data.up.sql

-- Insert default AI Models (Gemini + OpenAI)
INSERT INTO ai_models (id, name, provider, temperature, max_tokens, context_window, supports_tools, supports_stream, enabled)
VALUES 
('01H00000000000000000000001', 'gemini-3.1-flash-lite', 'gemini', 0.7, 4096, 1000000, FALSE, FALSE, TRUE),
('01H00000000000000000000002', 'gemini-2.0-flash-lite', 'gemini', 0.7, 4096, 1048576, FALSE, FALSE, TRUE),
('01H00000000000000000000003', 'gemini-embedding-2', 'gemini', 0.0, 0, 0, FALSE, FALSE, TRUE),
('01H00000000000000000000010', 'gpt-4o-mini', 'openai', 0.7, 4096, 128000, TRUE, TRUE, TRUE),
('01H00000000000000000000011', 'text-embedding-3-small', 'openai', 0.0, 0, 0, FALSE, FALSE, TRUE),
('01H00000000000000000000012', 'text-embedding-3-large', 'openai', 0.0, 0, 0, FALSE, FALSE, TRUE)
ON CONFLICT (name) DO UPDATE SET
    provider = EXCLUDED.provider,
    temperature = EXCLUDED.temperature,
    max_tokens = EXCLUDED.max_tokens,
    context_window = EXCLUDED.context_window,
    supports_tools = EXCLUDED.supports_tools,
    supports_stream = EXCLUDED.supports_stream,
    enabled = EXCLUDED.enabled;

-- Insert default Prompts
INSERT INTO prompts (id, name, system_prompt, description, model_id, active, version, created_at)
VALUES
(
    '01H00000000000000000000004', 
    'Knowledge Chunker', 
    'You are an expert knowledge extractor. Your task is to chunk the provided document into self-contained segments suitable for vector search. Return a JSON object with a "chunks" array. Each chunk must have "title", "content" (the detailed text), and "keywords" (array of strings).', 
    'Prompt used by the embedding worker to chunk knowledge documents into segments', 
    '01H00000000000000000000001', 
    TRUE, 
    1, 
    NOW()
),
(
    '01H00000000000000000000005', 
    'Default Assistant Prompt', 
    'You are Dani''s AI Portfolio Assistant. You help visitors learn about Dani''s experiences, projects, skills, and background. Be polite, concise, and helpful. Use the provided knowledge context when answering questions.', 
    'Default active prompt for general assistant chatbot sessions', 
    '01H00000000000000000000001', 
    TRUE, 
    1, 
    NOW() + INTERVAL '1 second'
),
(
    '01H00000000000000000000006', 
    'Memory Extractor', 
    'You are a memory extractor.
Your task is to analyze the conversation between the User and the Assistant and extract any useful facts or preferences about the User that would be valuable for future personalized conversations.

Ignore:
- greetings
- thanks
- jokes
- generic questions

Identify if there is any long-term memory to save.
The key should be a short, unique key/slug in lowercase with hyphens (e.g. "dan-ai-kafka" or "location-surabaya" or "favorite-language-golang").

Return a JSON object in this exact format:
{
  "save": true|false,
  "importance": 1-5,
  "category": "project|experience|certificate|skill|etc",
  "key": "short-unique-key-slug",
  "memory": "concise description of the visitor context (max 250 characters)"
}', 
    'Prompt used by the memory worker to extract facts from dialogue history', 
    '01H00000000000000000000001', 
    TRUE, 
    1, 
    NOW() + INTERVAL '2 seconds'
),
(
    '01H00000000000000000000007', 
    'Memory Consolidator', 
    'You are a memory consolidator. Your task is to combine two similar visitor memories into a single, cohesive memory without losing key context. Keep it short and concise (max 250 characters). Do not add metadata or introductions. Return only the consolidated memory string.', 
    'Prompt used by the memory worker to consolidate/merge duplicate or similar memories', 
    '01H00000000000000000000001', 
    TRUE, 
    1, 
    NOW() + INTERVAL '3 seconds'
)
ON CONFLICT (name) DO UPDATE SET
    system_prompt = EXCLUDED.system_prompt,
    description = EXCLUDED.description,
    model_id = EXCLUDED.model_id,
    active = EXCLUDED.active,
    version = EXCLUDED.version;

-- Insert active embedding profile (Gemini is default with 3072 dim)
INSERT INTO embedding_profiles (id, name, provider, model, dimension, metric_type, knowledge_collection, visitor_collection, enabled, created_at, updated_at)
VALUES (
    '01H00000000000000000000008',
    'e5',
    'gemini',
    'gemini-embedding-2',
    3072,
    'COSINE',
    'dan_knowledge_e5',
    'visitor_knowledge_e5',
    TRUE,
    NOW(),
    NOW()
),
(
    '01H00000000000000000000013',
    'openai-small',
    'openai',
    'text-embedding-3-small',
    1536,
    'COSINE',
    'dan_knowledge_openai',
    'visitor_knowledge_openai',
    TRUE,
    NOW(),
    NOW()
)
ON CONFLICT (name) DO UPDATE SET
    provider = EXCLUDED.provider,
    model = EXCLUDED.model,
    dimension = EXCLUDED.dimension,
    metric_type = EXCLUDED.metric_type,
    knowledge_collection = EXCLUDED.knowledge_collection,
    visitor_collection = EXCLUDED.visitor_collection,
    enabled = EXCLUDED.enabled,
    updated_at = NOW();

