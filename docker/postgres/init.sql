CREATE TABLE actors (
                        id UUID PRIMARY KEY,
                        tax_id TEXT NOT NULL,
                        type TEXT NOT NULL
);

CREATE TABLE accounts (
                          id UUID PRIMARY KEY,
                          dlt_account_id TEXT NOT NULL,
                          private_key TEXT NOT NULL,
                          public_key TEXT NOT NULL
);
