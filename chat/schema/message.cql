CREATE TABLE gochat.messages (
  channel_id uuid,
  message_id uuid,
  author_id bigint,
  content text,
  created_at timestamp,
  type text,
  PRIMARY KEY (channel_id, message_id)
);

create index on gochat.messages(type);
