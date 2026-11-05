#!/bin/bash
# Runs once, on first start with an empty data dir.
# Creates the chat-db app user that chat-service connects as.
set -e
mongosh --quiet -u "$MONGO_INITDB_ROOT_USERNAME" -p "$MONGO_INITDB_ROOT_PASSWORD" --authenticationDatabase admin <<JS
db = db.getSiblingDB("${MONGO_APP_DB:-chat-db}");
db.createUser({
  user: "${MONGO_APP_USER}",
  pwd: "${MONGO_APP_PASSWORD}",
  roles: [{ role: "readWrite", db: "${MONGO_APP_DB:-chat-db}" }]
});
JS
echo "Created Mongo user ${MONGO_APP_USER} on ${MONGO_APP_DB:-chat-db}"
