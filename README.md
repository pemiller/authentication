# Authentication
Authentication service using golang and couchbase

## Environment Variables

```
AUTHENTICATION_CB_CONNECTION=couchbase://username:password@localhost/bucket
AUTHENTICATION_PORT=9199
```

## Couchbase Indexes 

```n1ql
CREATE INDEX `idx_authentication_type`
ON `<bucket_name>`(__type)
```
