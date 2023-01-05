CREATE TABLE pipelines (
  id VARCHAR (255) PRIMARY KEY NOT NULL,
  user_id VARCHAR(255)
);

CREATE TABLE stages (
    id SERIAL PRIMARY KEY,
    pipeline_id VARCHAR(255) REFERENCES pipelines(id),
    name VARCHAR(255),
    message VARCHAR(255),
    status VARCHAR(16) CHECK (status IN ('SUCCESS', 'PENDING', 'RUNNING', 'FAILED'))
);