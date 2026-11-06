# ADR 0001: o iam no laboratório compartilhado

- Status: aprovado (2026-11-06)
- Escopo: ambiente `lab` apenas. Produção terá decisão própria.

## Contexto

O laboratório [aleogr/lab](https://github.com/aleogr/lab) mantém uma instância
Cloud SQL compartilhada (`lab-postgres`, Postgres 16, `db-f1-micro`,
`us-central1`) para os projetos em laboratório. Ela dorme de segunda a quinta,
das 22:00 às 07:30 (São Paulo), tem `max_connections` 25, dos quais 22 são dos
inquilinos, e o schooling já ocupa 8. A instância não será aumentada. O
aleogr/lab é dono da instância; cada inquilino declara o próprio banco, usuários
e grants no próprio repositório.

O projeto GCP anterior (`aleogr-identity-lab-6f2g`) foi excluído; seu id não
pode ser reutilizado.

## Decisão

1. **Projeto próprio** `aleogr-lab-iam-c4vc`, na mesma organização do
   `aleogr-lab-shared-dacd`, criado à mão pelo dono com faturamento, alerta de
   orçamento e bucket de state versionado.
2. **Terraform do iam aplicado pelo CI**, via Workload Identity Federation
   restrita ao id do repositório. Duas identidades: `iam-planner` (leitura, em
   PRs) e `iam-deployer` (escrita, só da `main`). A primeira aplicação é do dono.
3. **Contas de runtime:** `iam-release`, `iam-run`, `iam-migrate`.
4. **Banco como inquilino do `lab-postgres`:** banco `iam` com
   `deletion_policy = "ABANDON"`; `iam_migrator` (usuário com senha, dono do
   schema; a senha é gerada no Secret Manager e o usuário é criado à mão a
   partir dela e importado pelo Terraform, que nunca conhece a senha); `iam_app` (role sem login, só
   DML, sujeita a RLS); o serviço entra como usuário IAM `iam-run`, membro de
   `iam_app`, sem senha. Migração de bootstrap: `REVOKE CONNECT ... FROM PUBLIC`,
   `GRANT CONNECT` por nome, `CONNECTION LIMIT 6`.
5. **No aleogr/lab, por PR:** `tenants.iam = { declares = iam-deployer,
   connects = [iam-run, iam-migrate], logs_in = [iam-run],
   reads = [iam-release, iam-planner] }` (o planner lê o banco e os usuários do
   iam nos PRs; feito em aleogr/lab#17)
   e a linha na tabela "Who lives here".
6. **Teto de 6 conexões:** serviço com pool 2 × 1 instância × 2 revisões no
   rollout (4) + job de migração (1) + uma pessoa no psql (1). Tarefas agendadas
   chamam um endpoint do serviço e usam o pool dele (0). Sobram 8 na instância.
7. **Regras de código que sustentam o teto:** sem `LISTEN/NOTIFY` nem conexão
   dedicada (filas e outbox por polling); pool esgotado espera com timeout, nunca
   abre conexão extra; constante de orçamento com teste que falha se a soma
   passar de 6; deploy lê o `activationPolicy` e aborta com mensagem clara se a
   instância estiver parada; jobs agendados só depois das 08:00 em dias úteis.
8. **Ordem:** (1) projeto e bucket; (2) bootstrap do Terraform; (3) PR no
   aleogr/lab, só depois de as contas existirem; (4) banco, usuários e bootstrap
   SQL, quando o marco 1 precisar de banco no laboratório.

## Alternativas

- **Cloud SQL próprio:** isolamento total, mas custo fixo mensal sem receita
  (premissa 17).
- **Usuário com senha para o serviço:** mais simples, mas acrescenta um segredo
  de uso contínuo; o IAM elimina esse segredo.
- **Terraform aplicado pelo dono no Cloud Shell (como o schooling):** menos
  peças, mas foge do pipeline auditável e do padrão do brief (4.18).
- **Uma só identidade de CI:** um PR poderia aplicar a partir da própria branch.

## Consequências

- O iam do laboratório fica indisponível à noite de segunda a quinta e sujeito
  a um restore da instância inteira, que afeta também o schooling.
- `sqlTenant` não isola inquilinos: o iam nunca toca bancos, usuários ou roles
  que não sejam `iam`/`iam_*`.
- `iam-deployer` tem `projectIamAdmin` e é, na prática, dono do projeto; por
  isso só é alcançável pela `main`.
- Separar AuthN e AuthZ em serviços diferentes exige recalcular o teto.
- No dia em que o iam for para produção, ele sai do `lab-postgres`.
