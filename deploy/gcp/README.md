# GCP: ambiente de laboratório

Infraestrutura própria do iam no ambiente `lab`. O banco **não** fica aqui: o iam
é inquilino da instância compartilhada `lab-postgres`, que pertence ao
[aleogr/lab](https://github.com/aleogr/lab). Decisão: [ADR 0001](../../docs/adr/0001-lab-hosting.md).

| | |
|---|---|
| Projeto | `aleogr-lab-iam-c4vc`, na organização `421838980939` (a mesma do `aleogr-lab-shared-dacd`) |
| Bucket de state | `gs://aleogr-lab-iam-c4vc-tfstate` (versionado) |
| Quem aplica | o CI, via `.github/workflows/gcp.yml`, só a partir da `main` |

## Identidades

| Conta | Função |
|---|---|
| `iam-planner` | `terraform plan` em PRs (qualquer ref do repositório); só leitura |
| `iam-deployer` | `terraform apply` na `main`; na prática, dono do projeto |
| `iam-release` | publica imagens e revisões, roda o job de migração, lê o `activationPolicy` da instância |
| `iam-run` | executa o serviço; entra no banco como usuário IAM, sem senha |
| `iam-migrate` | executa o job de migração como `iam_migrator` |

A federação confia no **id** do repositório (`1443677386`), não no nome, para
sobreviver à transferência para a organização definitiva.

## O que foi feito à mão, e por quê

**Executado em 2026-11-06** pelo dono, no Cloud Shell, seções 1 a 4: projeto
criado e movido para a organização, faturamento vinculado, orçamento de R$ 25
(alertas em 50%, 90%, 100% e previsão de 100%), bucket criado e versionado,
primeira aplicação com 22 recursos, acesso ao state e variáveis do GitHub.
Número do projeto: `249031560608`.

Executado uma vez, pelo dono, no Cloud Shell. Nada disso é código porque o
Terraform não consegue criar o lugar onde guarda o próprio state, nem conceder a
si mesmo o acesso que precisa para começar.

### 1. Projeto, faturamento, orçamento e bucket

```sh
PROJECT=aleogr-lab-iam-c4vc
BILLING=XXXXXX-XXXXXX-XXXXXX   # gcloud billing accounts list
BUDGET=25BRL                   # moeda da conta de faturamento

gcloud projects create "$PROJECT" --name="iam lab" --organization=421838980939
gcloud billing projects link "$PROJECT" --billing-account="$BILLING"

gcloud services enable --project="$PROJECT" \
  serviceusage.googleapis.com cloudresourcemanager.googleapis.com \
  storage.googleapis.com billingbudgets.googleapis.com \
  iam.googleapis.com iamcredentials.googleapis.com sts.googleapis.com

gcloud billing budgets create --billing-project="$PROJECT" \
  --billing-account="$BILLING" --display-name="$PROJECT" \
  --budget-amount="$BUDGET" --filter-projects="projects/$PROJECT" \
  --threshold-rule=percent=0.5 --threshold-rule=percent=0.9 \
  --threshold-rule=percent=1.0 --threshold-rule=percent=1.0,basis=forecasted-spend

gcloud storage buckets create "gs://$PROJECT-tfstate" --project="$PROJECT" \
  --location=us-central1 --uniform-bucket-level-access --public-access-prevention
gcloud storage buckets update "gs://$PROJECT-tfstate" --versioning
```

As APIs também estão declaradas em `services.tf`; habilitá-las antes evita a corrida da primeira aplicação, em que contas de serviço seriam criadas antes da API do IAM responder.

O orçamento só avisa (e-mail aos administradores do faturamento); não desliga nada.

### 2. Primeira aplicação, pelo dono

A federação não pode ser aplicada por uma identidade que ainda não existe.

```sh
git clone --branch <branch-do-PR> https://github.com/aleogr/iam && cd iam/deploy/gcp/terraform
terraform version   # precisa ser 1.16.x (ver .terraform-version)
terraform init -backend-config="bucket=$PROJECT-tfstate" -backend-config=lab/backend.hcl
terraform apply -var="project_id=$PROJECT" -var-file=lab/lab.tfvars
terraform output
```

### 3. Acesso das identidades de CI ao state

Não declarado de propósito: uma configuração que gerencia o próprio acesso ao
state o revogaria no meio de um `destroy`.

```sh
gcloud storage buckets add-iam-policy-binding "gs://$PROJECT-tfstate" \
  --member="serviceAccount:iam-planner@$PROJECT.iam.gserviceaccount.com" \
  --role=roles/storage.objectViewer
gcloud storage buckets add-iam-policy-binding "gs://$PROJECT-tfstate" \
  --member="serviceAccount:iam-deployer@$PROJECT.iam.gserviceaccount.com" \
  --role=roles/storage.objectAdmin
```

### 4. GitHub

- **Settings → Actions → General → Allow specified actions**, exatamente:
  `google-github-actions/auth@7c6bc770dae815cd3e89ee6cdf493a5fab2cc093,`
  `hashicorp/setup-terraform@dfe3c3f87815947d99a8997f908cb6525fc44e9e`
- **Settings → Secrets and variables → Actions → Variables** (não são segredos):

| Variável | Valor |
|---|---|
| `GCP_PROJECT_ID` | `aleogr-lab-iam-c4vc` |
| `TF_STATE_BUCKET` | `aleogr-lab-iam-c4vc-tfstate` |
| `GCP_WIF_PROVIDER` | `terraform output -raw workload_identity_provider` |
| `GCP_PLANNER_SA` | `terraform output -raw planner_email` |
| `GCP_DEPLOYER_SA` | `terraform output -raw deployer_email` |

### 5. Banco no `lab-postgres`

Só com a instância acordada (fora de seg–qui, 22:00–07:30). Antes do merge do PR
que declara o banco, o dono habilita `sqladmin` e `secretmanager` e concede
`roles/secretmanager.admin` ao `iam-deployer` (a declaração em
`ci_identities.tf` mantém isso depois; conceder e usar no mesmo apply corre
contra a propagação do IAM). Depois do apply:

1. Gerar a senha direto no Secret Manager, sem passar pela tela:
   `openssl rand -base64 32 | tr -d '\n' | gcloud secrets versions add iam-migrator-password --project=$PROJECT --data-file=-`
2. Definir a senha do `iam_migrator` a partir do secret
   (`gcloud sql users set-password`).
3. Abrir o Cloud SQL Auth Proxy em `127.0.0.1` e rodar `../sql/bootstrap.sql`
   como `iam_migrator` no banco `iam`.
4. Conferir: `CONNECT` só para `iam_migrator` e `iam_app`, `datconnlimit = 6`,
   `iam-run` membro de `iam_app`.

Nunca usar `gcloud sql connect`: ele acrescenta uma rede autorizada à instância,
que pertence ao aleogr/lab.

## Ao acrescentar um recurso

Todo tipo de recurso novo precisa de concessão para as duas identidades de CI
(`ci_identities.tf`): leitura para o planner, escrita para o deployer, inclusive
`iam.serviceAccounts.actAs` sobre a conta com que o recurso roda. O bootstrap
roda como dono e esconde essas faltas; o primeiro run do CI é o primeiro teste
honesto. Mudanças de IAM levam alguns minutos para propagar.
