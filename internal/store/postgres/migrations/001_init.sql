CREATE TABLE IF NOT EXISTS projects (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    client TEXT NOT NULL CHECK (length(trim(client)) > 0),
    budget_kopecks BIGINT NOT NULL CHECK (budget_kopecks > 0),
    reserved_kopecks BIGINT NOT NULL DEFAULT 0 CHECK (reserved_kopecks >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (reserved_kopecks <= budget_kopecks)
);

CREATE TABLE IF NOT EXISTS payment_requests (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    purpose TEXT NOT NULL CHECK (length(trim(purpose)) > 0),
    amount_kopecks BIGINT NOT NULL CHECK (amount_kopecks > 0),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS payment_requests_project_id_idx
    ON payment_requests(project_id);
