# Amazon Verified Permissions (AVP)

| Campo | Valor |
|---|---|
| Categoria | Serviço gerenciado de autorização (PDP na nuvem AWS) baseado na linguagem Cedar |
| Site / repositório | https://aws.amazon.com/verified-permissions/ · guia: https://docs.aws.amazon.com/verifiedpermissions/latest/userguide/ · Cedar: https://github.com/cedar-policy/cedar · agente local: https://github.com/awslabs/avp-local-agent |
| Licença e modelo de negócio | Serviço proprietário, pago por chamada de API, sem mínimo. **Cedar** (linguagem, motor, validador) é Apache-2.0. Agente local `avp-local-agent` Apache-2.0, **experimental** |
| Stack | Cedar em Rust (com prova formal de partes do motor publicada pela AWS); API AWS (SigV4/IAM); integra com Cognito, OIDC, API Gateway (Lambda authorizer), CloudFormation; SDKs AWS em todas as linguagens; middleware Express |
| Maturidade | Preview em 11/2022, GA em 06/2023. Evolução lenta do guia: fontes OIDC (06/2024), batch com token (04/2024), aliases de policy store (02/2025), nomes de políticas (03/2025), múltiplos namespaces (05/2026), hard delete de alias (04/2026) |
| Acesso | 2026-11-03 |

## Resumo
- PDP gerenciado: a aplicação chama `IsAuthorized` (ou `IsAuthorizedWithToken`) com principal, ação, recurso, contexto e **lista de entidades**; o serviço avalia as políticas Cedar guardadas num *policy store*.
- O serviço guarda políticas, templates e esquema, **não os dados das entidades**: a aplicação envia atributos e hierarquias a cada pedido.
- Forte em análise de políticas (Cedar validável contra esquema), fraco em listagem, sincronização, avaliação local e ferramentas de governança.
- Preço por chamada torna o custo previsível em volume baixo e caro em volume alto sem cache.

## Funcionalidades por domínio

| Domínio (5.2) | Suporte | Notas |
|---|---|---|
| 1. Modelos | ◐ | RBAC (grupos como pais), ABAC (atributos, contexto), PBAC (Cedar `permit`/`forbid` com `when`/`unless`). ReBAC limitado a hierarquias enviadas no pedido (até 100 pais transitivos) e templates |
| 2. Decisões | ◐ | Single e batch (até 30 decisões por chamada). **Sem listagem** de recursos/usuários (padrão AWS: filtrar no banco e confirmar por batch). Resposta indica políticas determinantes e erros |
| 3. Políticas | ◐ | Cedar com esquema e modo de validação estrito; templates (40 por store) para atribuições por instância; nomes e aliases; IaC via CloudFormation. **Sem versionamento, simulação ou implantação gradual nativos** |
| 4. Dados | ○ | Entidades não são armazenadas: vêm no pedido ou do token (Cognito/OIDC). Sincronização é 100% problema da aplicação |
| 5. Acesso dinâmico | ◐ | Temporário via condição de tempo no contexto; templates para concessões pontuais. Sem pedidos/aprovações, M de N ou delegação |
| 6. Estados e transições | ○ | Estado só como atributo/contexto em condição |
| 7. Padrões | ◐ | AuthZEN apenas por **implementação de referência** publicada no blog da AWS (04/2025), não nativo. Tokens OIDC/OAuth como identidade |
| 8. Máquinas e agentes | ◐ | Principais de qualquer tipo; integração com API Gateway; nada específico para MCP encontrado |
| 9. Desempenho e resiliência | ◐ | 200 RPS por store em IsAuthorized e 30 RPS em batch (ajustáveis); cache por conta do cliente; agente local experimental em Rust |
| 10. Auditoria | ◐ | Integração AWS (CloudTrail para chamadas de gestão); log de decisões configurável não encontrado no guia consultado |
| 11. Experiência e administração | ◐ | Console AWS para policy stores, esquema, políticas, templates; assistente para API Gateway + Cognito. Sem UI para o cliente final |

## Diferenciais / exclusivos
- **Cedar**: linguagem pequena, analisável, com `forbid` prevalecendo, validação de tipos por esquema e base de verificação formal; aberta e reutilizável fora da AWS.
- **Policy templates** com slots `?principal`/`?resource`: concessões por instância (compartilhamento) sem gerar política nova à mão.
- **Identity sources**: transforma token Cognito/OIDC em principal e grupos automaticamente.
- Preço sem mínimo e sem servidor para operar.

## O que a comunidade elogia (com evidência)
- Separar autorização do código com linguagem analisável e aberta (cobertura de lançamento, DevClass 06/2023, imprensa).
- Custo zero de operação e integração nativa com API Gateway/Cognito (doc e blog AWS, fornecedor).
- Discussão independente relevante não encontrada (HN só com posts de divulgação; Reddit sem resultados na busca web).

## Do que a comunidade se queixa (com evidência)
- A própria AWS admite que casos com várias decisões por ação do usuário pedem batch e cache (blog "fine-grained authorization at scale"): latência de rede por decisão é o gargalo.
- Limite de 30 decisões por batch e 30 RPS de batch por store; filtragem de listas exige pré-filtrar no banco.
- Diferenças em relação ao Cedar nativo: templates exigem principal e recurso restritos; formato JSON de entidades diferente; esquema com chaves não vazias (guia AVP).
- Uma única identity source por policy store (cota não ajustável).
- Agente local marcado como experimental e com recomendação de atualização ≥ 15 s para não sofrer throttling.

## Histórico de segurança
- AVP (serviço): nenhum advisory público encontrado.
- Cedar (`cedar-policy`, crates.io):
  - **GHSA-m84h-6f26-p3p7 (2026, alta)**: esgotamento de memória ao analisar expressão `has` com caminho de atributos aninhado (DoS no parser).
  - **GHSA-c33x-vcv4-vg3g (2026, baixa)**: template JSON sem validação de slot permite pânico na vinculação e **divergência de decisão** entre representações.
- **Lições:** (1) parser de políticas é superfície de ataque (DoS por entrada maliciosa); limites de profundidade/tamanho e fuzzing contínuo; (2) toda representação alternativa da política (texto, JSON) deve passar pela mesma validação, com teste de equivalência de decisões; (3) mesmo motor com verificação formal teve falhas fora do núcleo verificado.

## Arquitetura e decisões notáveis
- Recurso central: *policy store* (políticas + templates + esquema + identity source). 30.000 stores por região por conta.
- Cotas relevantes: política até 10.000 bytes; esquema 100.000 bytes; 200.000 bytes de políticas por par principal-recurso (não ajustável); 100 pais transitivos por entidade (não ajustável).
- Preço (us-east-1, página oficial): single **US$ 0,000005/pedido** (US$ 5 por milhão); batch **US$ 0,00015/chamada** até 40 M/mês, US$ 0,000075 nos 60 M seguintes, US$ 0,00004 acima de 100 M; gestão de políticas US$ 0,00004/pedido.
- Sem estado de entidades no serviço: decisão é função pura de (políticas, pedido). Simplifica consistência, transfere a sincronização ao cliente.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Orientação oficial: **um policy store por tenant** (isolamento) ou store compartilhado com atributo de tenant comparado em toda política (exige atenção às cotas). Guia prescritivo dedicado.
- Extensibilidade: tipos e funções do Cedar apenas (sem funções externas).
- Configuração: console, API, CloudFormation.
- Soberania: dados de entidades não ficam na AWS (vêm no pedido); políticas ficam na região escolhida. Lock-in em AWS para o PDP, mitigado pelo Cedar ser aberto.

## Modelos, linguagem, consistência, listagem, sincronização, explicabilidade, testes, avaliação local, desempenho, AuthZEN
- **Linguagem:** Cedar (RBAC/ABAC/PBAC; hierarquias `in`; `forbid` prevalece; esquema tipado).
- **Consistência:** como não há dados de entidades no serviço, a decisão reflete o que a aplicação enviou; mudanças de política são eventualmente consistentes (não documentado em detalhe).
- **Listagem:** ausente; padrão recomendado é pré-filtro no banco + batch de 30.
- **Sincronização:** não há; a aplicação monta a lista de entidades por pedido (ou usa token).
- **Explicabilidade:** políticas determinantes na resposta.
- **Testes:** validação de esquema; testes de decisão ficam com o cliente (Cedar CLI/SDK).
- **Avaliação local:** `avp-local-agent` (Rust, experimental, v3.0.0 de 04/2025) usa `cedar-local-agent` com cache de políticas puxadas da AVP.
- **Desempenho publicado:** não encontrado número oficial de latência; só cotas de RPS.
- **AuthZEN:** referência externa, não nativo.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N):** **não.**
- **Autorização sensível ao estado / transições declarativas:** **não** como recurso; **parcial** como padrão: estado do recurso pode ir no pedido e políticas Cedar podem condicionar a ação (`when { resource.status == "draft" }`), o que é o "nível 1" da ideia (a aplicação informa o estado), mas sem declarar transições.

## Implicações para o projeto
1. **Adotar** o modelo "a aplicação envia o estado e as entidades relevantes no pedido" como opção (coerente com a decisão do brief de o AuthZ não sincronizar estado de negócio).
2. **Considerar Cedar** como linguagem de políticas embutível (Rust; avaliar binding Go ou motor próprio compatível): analisável, aberta, com `forbid` explícito.
3. **Adotar** templates com slots para concessões por instância (compartilhamento, acesso temporário) sem explosão de políticas.
4. **Evitar** cobrar ou limitar por decisão: o modelo por chamada empurra o cliente a caches inseguros; nossa avaliação local deve ser primeira classe.
5. **Oportunidade**: listagem, versionamento, simulação e logs de decisão são lacunas claras da AVP.
6. **Adotar** limites rígidos documentados (tamanho de política, profundidade de hierarquia) como defesa contra DoS, com alertas antes de atingir (premissa 4).
7. **Adotar** a opção de store por tenant como forma de isolamento forte em clientes multi-tenant.

## Fontes
- https://aws.amazon.com/verified-permissions/pricing/ — 2026-11-03 — fornecedor
- https://docs.aws.amazon.com/verifiedpermissions/latest/userguide/what-is-avp.html — 2026-11-03 — fornecedor
- https://docs.aws.amazon.com/verifiedpermissions/latest/userguide/doc-history.html — 2026-11-03 — fornecedor
- https://docs.aws.amazon.com/verifiedpermissions/latest/userguide/quotas.html — 2026-11-03 — fornecedor
- https://docs.aws.amazon.com/verifiedpermissions/latest/userguide/policy-validation-mode.html — 2026-11-03 — fornecedor
- https://docs.aws.amazon.com/verifiedpermissions/latest/userguide/policy-templates.html — 2026-11-03 — fornecedor
- https://docs.aws.amazon.com/verifiedpermissions/latest/userguide/terminology-differences-avp-cedar.html — 2026-11-03 — fornecedor
- https://docs.aws.amazon.com/verifiedpermissions/latest/userguide/policy-stores.html — 2026-11-03 — fornecedor
- https://docs.aws.amazon.com/prescriptive-guidance/latest/saas-multitenant-api-access-authorization/avp-design-per-tenant-store.html — 2026-11-03 — fornecedor (via busca)
- https://docs.aws.amazon.com/prescriptive-guidance/latest/saas-multitenant-api-access-authorization/avp-design-shared-store.html — 2026-11-03 — fornecedor (via busca)
- https://aws.amazon.com/blogs/security/use-amazon-verified-permissions-for-fine-grained-authorization-at-scale/ — 2026-11-03 — fornecedor
- https://aws.amazon.com/blogs/security/how-to-support-openid-authzen-requests-with-amazon-verified-permissions/ — 2026-11-03 — fornecedor (15/04/2025)
- https://github.com/awslabs/avp-local-agent e https://lib.rs/crates/avp-local-agent — 2026-11-03 — fornecedor (via busca; experimental, v3.0.0)
- https://github.com/cedar-policy/cedar/security/advisories — 2026-11-03 — advisories
- https://api.osv.dev/v1/query (crates.io cedar-policy) — 2026-11-03 — base neutra
- https://devclass.com/2023/06/14/aws-verified-permissions-for-custom-apps-now-generally-available-built-on-open-source-cedar-project — 2026-11-03 — imprensa (via busca)
- https://news.ycombinator.com/item?id=33783560 (anúncio do preview) — 2026-11-03 — comunidade
