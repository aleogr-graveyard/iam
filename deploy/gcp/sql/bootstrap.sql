-- Run ONCE by the owner, connected to the `iam` database as iam_migrator,
-- through the Cloud SQL Auth Proxy (deploy/gcp/README.md, section 5).
--
-- Only `iam` objects. The ceiling of 6 is docs/adr/0001-lab-hosting.md and
-- aleogr/lab's docs/lab.md; changing it means changing all three.

BEGIN;

-- Every role may connect to every database by default.
REVOKE CONNECT ON DATABASE iam FROM PUBLIC;

-- The worst case read from the design: pool 2 x 1 instance x 2 revisions,
-- one migration job, one person.
ALTER DATABASE iam CONNECTION LIMIT 6;

-- Data access only. Not the owner of any table, so row-level security
-- applies to it.
CREATE ROLE iam_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS;

GRANT CONNECT ON DATABASE iam TO iam_migrator, iam_app;

-- The service's IAM user takes its privileges from iam_app.
GRANT iam_app TO "iam-run@aleogr-lab-iam-c4vc.iam";

COMMIT;
