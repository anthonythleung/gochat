CREATE KEYSPACE chat_log WITH REPLICATION = { 'class': 'SimpleStrategy', 'replication_factor': 1};

CREATE TABLE chat_log.messages (
  channel_id uuid,
  message_id uuid,
  created_at timestamp,
  author_id bigint,
  content text,
  type text,
  PRIMARY KEY (channel_id, message_id)
); 
