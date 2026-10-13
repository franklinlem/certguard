# ADR-011 — Compose exclusivo de produção

Status: proposta implementada, implantação pendente.

## Contexto

O Compose Raw existente fixa domínio e nomes Traefik de homologação. Reutilizá-lo em produção poderia manter origem incorreta ou substituir rotas de homologação no proxy compartilhado.

## Decisão

Criar `compose.coolify.prod.yaml` com domínio/origem e nomes Traefik próprios. Manter UUID nativo do recurso Coolify como identidade Compose, segredo gerado no recurso novo e volume próprio. Somente web entra na rede compartilhada do proxy; não publicar portas de API/banco. Manter flags operacionais desligadas até validação específica. Não copiar dados ou segredos de homologação.

A CI renderiza os dois arquivos com identidades padrão e UUIDs simulados; a mesma política valida isolamento e roteamento de cada ambiente. Pequena duplicação explícita facilita revisão do domínio/rotas e evita alteração do ambiente já implantado. Mudanças comuns futuras devem ser aplicadas nos dois arquivos e validadas na matriz.

## Consequências

A produção poderá ser provisionada sem mudanças globais de proxy. Criar o arquivo não conclui homologação, backup, restauração ou autorização para mudanças com risco de interrupção/perda de dados. O rollback preserva volume e schema.
