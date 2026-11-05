# Aserto / Topaz

| Campo | Valor |
|---|---|
| Categoria | Topaz: serviço de autorização open source (OPA + diretório estilo Zanzibar), implantado ao lado da aplicação. Aserto: plano de controle SaaS sobre Topaz (**encerrado**) |
| Site / repositório | https://www.topaz.sh · https://github.com/aserto-dev/topaz · (histórico) https://www.aserto.com |
| Licença e modelo de negócio | Topaz Apache-2.0. **A Aserto encerrou as operações em 31/05/2025**: o SaaS de controle foi desligado e a empresa dissolvida porque "não conseguimos fazer a economia desse modelo fechar". Mantenedores seguem com Topaz e "podem oferecer suporte comercial" (maintainers@topaz.sh). Não há mais produto pago oficial |
| Stack | Go; OPA embutido como motor de decisão (Rego); diretório em **BoltDB** embutido; APIs gRPC e REST; políticas empacotadas como imagens OCI; SDKs Node, Go, Python, Java, .NET, Ruby, React |
| Maturidade | Lançado em 10/2022 (Show HN, 132 pontos). ~1,4 mil estrelas, 49 forks, 9 issues abertas, ~920 commits. Ainda em 0.x: **v0.33.25 em 26/10/2026**; lançamentos de 09–10/2026 são majoritariamente atualização de OPA e dependências |
| Acesso | 2026-11-03 |

## Resumo
- Combina políticas Rego (OPA) com um diretório de objetos e relações (ReBAC estilo Zanzibar) consultado por built-ins (`ds.check`, `ds.object`...) dentro da política.
- Projetado para rodar local (sidecar ou microserviço) com dados em banco embutido: decisões em ~1 ms segundo o fornecedor.
- Um dos co-chairs do grupo AuthZEN; Topaz foi implementação de referência, com avaliação, lote e buscas.
- A empresa que o sustentava fechou em 2025; o projeto continua vivo, mas em modo manutenção, com equipe pequena.

## Funcionalidades por domínio

| Domínio (5.2) | Suporte | Notas |
|---|---|---|
| 1. Modelos | ● | RBAC, ABAC (atributos de usuário, recurso e ambiente em Rego) e ReBAC (manifesto de tipos, relações, permissões), combináveis na mesma política |
| 2. Decisões | ● | `is` (decisão), `decisiontree` (mapa de decisões para a UI), `query` (Rego arbitrário); diretório: `Check`, `GetGraph` (busca de sujeitos/objetos); AuthZEN evaluation(s) e search. Explicabilidade limitada a decision logs |
| 3. Políticas | ◐ | Rego em Git; **imagens OCI de política** com build, tag, push, pull e **assinatura** (`policy sign`); testes por asserções JSON (`topaz directory test`, `topaz authorizer test`). Sem simulação/implantação gradual nativas |
| 4. Dados | ◐ | Diretório local (objetos, relações, manifesto); carga por `ds-load` com plugins (Auth0, Azure AD, Cognito, Google, LDAP, Okta, custom). Sem sincronização contínua após o fim do Aserto; consistência local |
| 5. Acesso dinâmico | ○ | Nada de acesso temporário, solicitações, aprovações ou delegação nativos (expressáveis em Rego com dados) |
| 6. Estados e transições | ○ | Não |
| 7. Padrões | ● | AuthZEN nativo desde 0.32.54 (HTTP e gRPC), inclusive buscas de sujeito/recurso/ação (draft 03). Sem RAR/UMA |
| 8. Máquinas e agentes | ◐ | Identidades de máquina no diretório; nada específico para agentes/MCP |
| 9. Desempenho e resiliência | ● | Sidecar/microserviço com dados locais; carga sob demanda de objetos do diretório (não materializa tudo em memória como `data.json`) |
| 10. Auditoria | ◐ | Decision logs (plugin OPA) em arquivos rotativos; agregação/central era do Aserto |
| 11. Experiência e administração | ◐ | Console web local do Topaz, CLI rica, templates; UI central multi-instância se foi com o Aserto |

## Diferenciais / exclusivos
- **OPA + diretório Zanzibar no mesmo processo**: política declarativa (PBAC/ABAC) que consulta relações sem chamada de rede.
- **Políticas como artefatos OCI assinados**: versionamento imutável, distribuição por registry, cadeia de suprimentos verificável.
- **Display state map / decisiontree**: uma chamada devolve `visible`/`enabled` para cada rota, para a UI esconder ou desabilitar ações.
- AuthZEN completo cedo (com buscas), influência direta no padrão.

## O que a comunidade elogia (com evidência)
- Combinação de policy-as-code com ReBAC permitindo RBAC/ABAC/ReBAC intercambiáveis (Show HN 2022; comentário de ex-funcionário, viés).
- Boas-vindas de mantenedores de OpenFGA e SpiceDB ao novo projeto aberto Zanzibar (Show HN 2022, comunidade/concorrentes).
- Disponibilidade como bloco aberto em vez de construir internamente (Show HN 2022).

## Do que a comunidade se queixa (com evidência)
- Latência de chamada a serviço externo se acumula por requisição; preferência por biblioteca importável (Show HN 2022).
- Colisão de nome "Topaz" com outros produtos (Show HN 2022).
- Issues mais comentadas: erro de certificado gRPC na carga de dados (#50), instalação/permissões (#23), pull da imagem de política (#120), dúvidas de escalabilidade do diretório (#243, #135), backup BoltDB travando com o servidor ativo (#674), **pedido de operações em lote transacionais (#667, aberta desde 02/2026)**.
- Encerramento do Aserto: clientes do SaaS perderam o plano de controle (sincronização de borda, gestão central) com cerca de um mês de aviso (post de 27/04/2025, desligamento em 31/05/2025).

## Histórico de segurança
- **GHSA-82cq-px67-wrmp (2026, crítica, CVSS 9.1, CWE-287)**: o autorizador escolhia o JWKS a partir do `iss` **não verificado** do JWT; sem allowlist de emissores, qualquer um podia hospedar discovery próprio e forjar identidade de qualquer usuário. Efeitos secundários: SSRF (busca URLs indicadas pelo atacante) e "chute" de `<iss>/.well-known/jwks.json` quando o discovery falhava. Corrigido em 0.33.16; sem CVE atribuído no advisory.
- OSV: nenhuma outra vulnerabilidade para `github.com/aserto-dev/topaz` e `go-directory`.
- SECURITY.md: só a última versão recebe correções até a 1.x; relato privado pelo GitHub, resposta em 48 h.
- **Lições:** (1) num PDP que aceita identidade por token, emissores e JWKS devem ser **configuração explícita** (allowlist), nunca derivados do token; (2) validar `iss` e `aud` antes de qualquer busca de rede; (3) discovery só HTTPS e hosts configurados; (4) falha de resolução de chave = negar (fail closed), nunca adivinhar URL. Entra direto no modelo de ameaças do AuthZ (premissa 8: IdPs de terceiros).

## Arquitetura e decisões notáveis
- Autorizador (OPA + built-ins) + diretório (BoltDB) + decision logger + console, num binário/contêiner.
- Objetos do diretório carregados sob demanda pelos built-ins, evitando o modelo "tudo em memória" do OPA.
- Política resolvida por imagem OCI (ex.: `ghcr.io/...:latest`), configurada no arquivo do Topaz.
- O modelo de negócio do Aserto era **híbrido** (controle SaaS + autorizadores no cliente com sincronização em tempo real); os fundadores concluíram que não fechava economicamente.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy modelada no manifesto/política (tenant como objeto); site fala em "evoluir de RBAC multi-tenant para ABAC/ReBAC". Não há isolamento por tenant no motor (uma instância por conjunto de dados).
- Banco: só BoltDB embutido (sem Postgres/MySQL); backup por ferramenta própria.
- Configuração por arquivo YAML e CLI (`topaz config`, templates).
- Soberania: total; tudo roda na infraestrutura do cliente.

## Modelos, linguagem, consistência, listagem, sincronização, explicabilidade, testes, avaliação local, desempenho, AuthZEN
- **Linguagem:** Rego + manifesto YAML do diretório (tipos, relações, permissões com união/interseção/exclusão).
- **Consistência:** forte local (um nó, um banco embutido); entre réplicas, não há mecanismo após o fim do Aserto.
- **Listagem:** `GetGraph` e buscas AuthZEN (sujeitos que podem X no recurso; recursos que o sujeito pode X).
- **Sincronização:** carga em lote/incremental via `ds-load` e API de escrita do diretório; sem CDC ou eventos nativos; a aplicação faz a escrita dupla.
- **Explicabilidade:** decision logs com contexto; sem "por que" estruturado.
- **Testes:** asserções JSON para diretório e para decisões do autorizador.
- **Avaliação local:** é o modelo padrão (sidecar).
- **Desempenho publicado:** "~1 ms" (site do Aserto, marketing); sem benchmark metodológico encontrado.
- **AuthZEN:** sim (evaluation, evaluations, search).

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N):** **não.** Nenhum recurso de solicitação/aprovação.
- **Autorização sensível ao estado / transições declarativas:** **não.** Estado só como atributo usado em Rego.

## Implicações para o projeto
1. **Adotar** políticas como artefatos versionados, imutáveis e assinados (OCI ou equivalente), verificados pelo PDP antes de carregar: atende versionamento, governança e auditoria do próprio software (premissa 5).
2. **Adotar** endpoint de "mapa de decisões para a UI" (visible/enabled por ação) como recurso do SDK.
3. **Adotar** a combinação política declarativa + grafo de relações no mesmo processo, que preserva o monólito modular e o custo baixo.
4. **Evitar** o erro do GHSA-82cq: emissores confiáveis e JWKS fixados por tenant/aplicação; teste ofensivo automatizado para "token com iss do atacante".
5. **Evitar** depender de um plano de controle SaaS para funções essenciais (sync, gestão): a edição self-hosted precisa ser completa.
6. **Lição de negócio**: o modelo híbrido "SaaS controla PDPs do cliente" não se pagou para a Aserto; monetização do projeto deve considerar suporte, hospedagem gerenciada completa ou recursos corporativos, não só o plano de controle.
7. **Oportunidade**: operações em lote transacionais (#667) e backup online do armazenamento são lacunas pedidas.

## Fontes
- https://www.aserto.com/contact-engineer — 2026-11-03 — fornecedor ("Aserto has ceased operations")
- https://www.aserto.com/blog/the-final-chapter-for-aserto — 2026-11-03 — fornecedor (motivos do encerramento)
- https://www.aserto.com/ — 2026-11-03 — fornecedor (marketing, "~1ms")
- https://www.aserto.com/blog/authzen-gartner-iam-and-native-support-in-topaz — 2026-11-03 — fornecedor
- https://www.topaz.sh/ — 2026-11-03 — projeto
- https://www.topaz.sh/docs/architecture — 2026-11-03 — projeto
- https://www.topaz.sh/docs/authorizer-guide/decisiontree — 2026-11-03 — projeto
- https://www.topaz.sh/docs/authorizer-guide/display-state-map — 2026-11-03 — projeto
- https://www.topaz.sh/docs/directory/model — 2026-11-03 — projeto
- https://www.topaz.sh/docs/directory/reader — 2026-11-03 — projeto
- https://www.topaz.sh/docs/decision-logs — 2026-11-03 — projeto
- https://www.topaz.sh/docs/policies/lifecycle — 2026-11-03 — projeto
- https://www.topaz.sh/docs/command-line-interface/topaz-cli/test — 2026-11-03 — projeto
- https://www.topaz.sh/sitemap.xml (plugins ds-load, implantação sidecar/helm/ECS) — 2026-11-03 — projeto
- https://github.com/aserto-dev/topaz — 2026-11-03 — repositório
- https://github.com/aserto-dev/topaz/releases — 2026-11-03 — repositório
- https://github.com/aserto-dev/topaz/issues?q=is%3Aissue+sort%3Acomments-desc — 2026-11-03 — comunidade
- https://github.com/aserto-dev/topaz/security — 2026-11-03 — SECURITY.md
- https://github.com/aserto-dev/topaz/security/advisories/GHSA-82cq-px67-wrmp — 2026-11-03 — advisory
- https://news.ycombinator.com/item?id=33317597 (Show HN Topaz, 2022) — 2026-11-03 — comunidade (inclui concorrentes)
- https://news.ycombinator.com/item?id=38164327 (Show HN Topaz 0.30, 2023) — 2026-11-03 — comunidade
- https://pitchbook.com/profiles/company/461793-70 — 2026-11-03 — terceiro (via busca: "Out of Business 31-May-2025")
- https://api.osv.dev/v1/query (topaz, go-directory) — 2026-11-03 — base neutra
