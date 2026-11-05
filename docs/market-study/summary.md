# Estudo de mercado preliminar — síntese

Período: 2026-11-03 a 2026-11-05. 32 players de AuthN, 24 de AuthZ, padrões e nomes.
Matrizes: [AuthN](matrix-authn.md) · [AuthZ](matrix-authz.md). Diferenciais: [differentials-check](authz/differentials-check.md).
Padrões: [AuthZEN](standards/authzen.md) · [outros de AuthZ](standards/authz-standards.md) ·
[segurança e regulação](standards/security-references.md) · [agentes e MCP](standards/agents-mcp.md) · [nomes](naming.md).

**Limitações:** vários sites oficiais e a API do GitHub ficaram inacessíveis em parte da pesquisa;
agentes recorreram a repositórios, curl, OSV, HN e buscas. O que veio só de busca ou de conhecimento
prévio está marcado nos documentos. Bases de marcas (INPI, EUIPO, USPTO) não foram consultadas.

## 1. Mínimo esperado

**AuthN**
- OIDC/OAuth 2.x certificado, PKCE, client credentials, device flow; passkeys, TOTP, WebAuthn.
- Organizações B2B: usuário em várias organizações, papel por membership, SSO e SCIM por organização,
  descoberta por domínio verificado + JIT, portal de autoatendimento para o TI do cliente.
- Barra de protocolos (Keycloak, Apache-2.0): FAPI 2.0, PAR, DPoP, CIBA, SCIM, OID4VCI/VP, AuthZEN experimental.
- Importação de hashes de senha e migração preguiçosa (FusionAuth, Cognito triggers).

**AuthZ**
- Check individual e em lote; ao menos uma forma de listagem; motivo da negação.
- Avaliação local ou sidecar com política vinda de um ponto central; logs de decisão.
- RBAC hierárquico por tenant; ReBAC (Zanzibar) ou ABAC (políticas), raramente ambos bem.
- AuthZEN 1.0 (final em 12/01/2026) começa a ser o contrato esperado (Cerbos nativo; PlainID,
  Axiomatics; OpenFGA e Keycloak experimentais).

## 2. Diferenciais que valem considerar

| Diferencial | Situação no mercado | Avaliação |
|---|---|---|
| Estados e transições declarativos (definição única para AuthZ e aplicação) | Ninguém tem; estado só como atributo | **Diferencial confirmado**, mas copiável; valor na definição única + checagem de transição legal |
| M de N com assinatura WebAuthn vinculada à ação exata, para qualquer aplicação | Quórum existe em PAM/nuvem (Teleport, Vault, GCP PAM, AWS MPA); assinatura por ação + quórum só na Turnkey (cripto) | **Diferencial parcial**: a combinação é inédita (serviço genérico + evidência verificável pelo executor + perfil AuthZEN de aprovação + aberto). Limite: WebAuthn não prova que o autenticador exibiu a ação |
| "Transição exige M de N assinado" | Não encontrado | Combinação dos dois acima; forte para o cenário da clínica |
| Sincronização de dados sem escrita dupla (AuthZ participando da transação da aplicação, ou listagem compilada em SQL sem copiar dados) | Ninguém resolve no código aberto (pedido aberto no OpenFGA; Oso Local Authorization fechou) | **Oportunidade direta** |
| Listagem eficiente e paginada | Ponto fraco de todos os motores Zanzibar | Oportunidade |
| AuthN + AuthZ irmãos, com o mesmo modelo de tenants | IdPs só põem papéis no token (Entra perde o claim acima de 200 grupos); clientes acabam com dois produtos | Oportunidade de posicionamento |
| Pisos de segurança herdados | Descope: política de baixo só endurece; usuário em várias organizações recebe a mais restritiva | Validação do brief 4.5 |
| Políticas como código revisável em PR, com replay de decisões reais antes de aplicar | Permit, Oso Migrate | Adotar |
| Export, backup e portabilidade completos | Cognito não exporta hashes; Auth0 só via suporte | Oportunidade |
| Multi-tenancy, UI administrativa e correções de segurança completos no código aberto | Ory, SuperTokens, Logto, Pomerium cobram por isso | Posicionamento |

## 3. Queixas recorrentes (oportunidades)

- Segurança como add-on pago: senhas vazadas, bot, MFA adaptativo, SAML, SCIM, patches com SLA.
- Preço por MAU e por conexão SSO/SCIM (~US$ 125/mês cada no WorkOS e no Stytch).
- Peso e complexidade operacional (Keycloak/JVM, Authentik/Django, Ory headless).
- Falta de residência de dados na UE e no Brasil (exceção: Descope, com São Paulo).
- Lock-in (export) e configurações imutáveis (Cognito).
- **Volatilidade de fornecedores e licenças:** Aserto encerrou (05/2025); Oso arquivou a biblioteca (10/2026);
  WorkOS FGA v1 descontinuado (11/2025); Permify Apache → AGPL e vendido à FusionAuth; Zitadel → AGPL;
  Pomerium com licença restritiva nos binários (11/2026); PocketBase saiu do GitHub; Axiomatics vendida
  a grupo de defesa; criadores do OPA foram para a Apple.
- Listagem limitada ou truncada sem aviso (OpenFGA: 3 s / 1000 itens).

## 4. Lições de segurança (insumo para o threat model)

| Classe | Exemplos | Regra para nós |
|---|---|---|
| Vinculação de contas | ATO por auto-link com e-mail não verificado (Better Auth, SuperTokens, PocketBase); dezenas de ATO na Zitadel em 2026 (vínculo com IdP, cadastro de passkey, device code) | Vincular só com e-mail verificado pelo IdP confiável; testes de pre-hijacking |
| Isolamento entre tenants | Entra "actor tokens" (CVSS 10, sem log); API key de tenant alcançando endpoints globais (FusionAuth, SuperTokens); passkey aceita em outro tenant | Suíte obrigatória de testes de isolamento em toda porta e todo endpoint |
| SAML | 6 CVEs no Authentik em 2026 (XSW, comentário no NameID, replay); bypass de assinatura em libs Go | Parser e validação auditados; replay cache; testes de XSW |
| Parsing e canonicalização | 5 bypasses no oauth2-proxy; Authelia; Oathkeeper CVSS 10; Keycloak (caminho não normalizado → regra ampla); Pomerium (`encoding/json` sem distinção de maiúsculas) | Parser canônico estrito, falhar fechado, testes diferenciais |
| Confiança em chaves e emissores | Topaz escolhia JWKS pelo `iss` não verificado (CVSS 9.1) | Emissores e JWKS em allowlist; falha de chave nega |
| Segredos e defaults | Kratos e Keto com segredo padrão no código (Keto: SQLi via cursor forjado); Better Auth anunciava `alg=none` e PKCE `plain` | Nenhum segredo padrão; cursores e tokens autenticados; padrões seguros |
| Uso único e concorrência | Código de autorização resgatado duas vezes; família de refresh bifurcada (Better Auth) | Operações atômicas no banco + testes de concorrência |
| Coerência da autorização | 29 advisories no OpenFGA (bordas: exclusão, wildcard, userset; fail-open em subárvore); Cerbos e SpiceDB com listagem mais permissiva que o check; Keycloak lote checando só o 1º recurso | Testes diferenciais check × list × lote; negar em erro no meio da avaliação |
| Relógio e entrada do cliente | Keycloak aceitava claim de hora do cliente | Relógio só do servidor (injetado) |
| Código do cliente no servidor | RCE em políticas Python (Authentik), DoS em VM JS (Logto) | Extensões fora do processo, com limites de tempo e memória |
| DoS em parsers | 8 de 12 vulns em libs Go/Rust; go-jose com 5 em 2026; parsers recursivos SCIM/LDAP no Kanidm | Limites de tamanho e profundidade em toda entrada |
| SDKs e artefatos | Clerk (predicados de autorização), WorkOS (token no HTML, CDN servindo sessão alheia, refresh em log), Okta (HAR com tokens) | `no-store`, redação de tokens em logs e artefatos, SDKs testados como código crítico |
| Suporte e recuperação | Engenharia social em helpdesk para resetar MFA (Okta); bcrypt truncando em 72 bytes | Recuperação nunca mais fraca que o MFA; impersonação auditada; nunca bcrypt como KDF genérica |
| Elevação de privilégio | Teleport RFD 0131 | MFA por requisição, nunca sessão elevada reaproveitável |

## 5. Implicações para o brief

**Confirma**
- Contratos primeiro com AuthZEN; M de N e estados/transições como diferenciais (4.29).
- Promover organização a tenant (4.7); pisos de segurança (4.5); exceção de segurança para dependências (4.30).
- Self-hosted completo, sem recursos de segurança pagos, como posicionamento (seção 8).

**Muda ou acrescenta**
- **Event sourcing [ABERTO, 4.4]:** a Zitadel está saindo dele (consistência eventual, projeções travadas,
  consultas lentas); a próxima versão usa tabelas relacionais como verdade e eventos só para auditoria.
  Recomendação: modelo de estado tradicional + log de auditoria encadeado.
- **Motor de AuthZ [ABERTO, 4.29]:** nenhum motor existente cobre ReBAC + ABAC + estados + listagem +
  consistência, e os fornecedores são instáveis. Candidatos a embutir: `cedar-go` (3 dependências, mais
  rápido, verificação formal, mas sem avaliação parcial) ou OpenFGA embutido (mais completo em ReBAC,
  mais dependências, histórico de bypass). Tema central da discussão de AuthZ.
- **Dependências Go:** stdlib + `x/crypto`, go-jose v4 (≥ 4.1.8), go-webauthn. Estender a exceção da
  premissa 24 a JOSE/JWT. Nenhuma lib Go de identidade tem auditoria independente publicada.
- **MFA do Spring Security 7:** cada fator como autoridade com carimbo de tempo; a política exige fatores
  e frescor. Encaixa na sessão central do brief (4.5).
- **NIST SP 800-63-4:** senha ≥ 15 caracteres como fator único; sem regras de composição nem troca
  periódica; passkeys sincronizáveis só até AAL2; AAL2 precisa oferecer opção resistente a phishing.
- **Cyber Resilience Act:** gestão de identidade e acesso no Anexo III, Classe I. Com monetização, o
  projeto é fabricante; notificação de vulnerabilidades já vale desde 11/09/2026.
- **Agentes e MCP:** especificação do MCP 2026-07-28 (Client ID Metadata Documents, ID-JAG); entra no
  domínio "máquinas e agentes" dos dois produtos.
- **Nomes:** todos os 70 candidatos têm `.com` registrado; mais limpos: Agnosco/Permitto, Ipse/Licet;
  plataforma Gatehall. Evitar Custos (IAM da Apache). `iam` só como nome interno.
- **Lacunas do mercado a explorar:** ICP-Brasil, WhatsApp, LDAP direto e direitos do titular (LGPD) quase
  inexistentes.
