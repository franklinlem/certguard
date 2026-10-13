# Revisão de PRs — 13/10/2026

Leitura do GitHub, sem merges/fechamentos ou mudanças na VPS. PRs de desenvolvimento #6–#10 sem conflitos, revisões ou comentários de alteração pendentes.

| PR | Entrega | Base | Última CI |
|---|---|---|---|
| [#1](https://github.com/franklinlem/certguard/pull/1) | build(deps): bump actions/setup-node from 4.4.0 to 7.1.0 | main | [Segurança: success](https://github.com/franklinlem/certguard/actions/runs/38291338439), [Validação: success](https://github.com/franklinlem/certguard/actions/runs/38291338412) |
| [#2](https://github.com/franklinlem/certguard/pull/2) | build(deps): bump alpine from 3.22 to 3.24 in /api | main | [Segurança: success](https://github.com/franklinlem/certguard/actions/runs/38291338682), [Validação: success](https://github.com/franklinlem/certguard/actions/runs/38291338584) |
| [#3](https://github.com/franklinlem/certguard/pull/3) | build(deps): bump actions/setup-go from 5.6.0 to 7.0.0 | main | [Segurança: success](https://github.com/franklinlem/certguard/actions/runs/38291349492), [Validação: success](https://github.com/franklinlem/certguard/actions/runs/38291349693) |
| [#4](https://github.com/franklinlem/certguard/pull/4) | build(deps): bump actions/checkout from 4.4.0 to 7.0.1 | main | [Segurança: success](https://github.com/franklinlem/certguard/actions/runs/38291356689), [Validação: success](https://github.com/franklinlem/certguard/actions/runs/38291356768) |
| [#5](https://github.com/franklinlem/certguard/pull/5) | build(deps): bump node from 22.23.2-alpine to 26.11.1-alpine in /web | main | [Validação: success](https://github.com/franklinlem/certguard/actions/runs/38291540881), [Segurança: success](https://github.com/franklinlem/certguard/actions/runs/38291540844) |
| [#6](https://github.com/franklinlem/certguard/pull/6) | Adicionar autenticação, endpoints e verificações TLS persistidas | main | [Segurança: success](https://github.com/franklinlem/certguard/actions/runs/38294848853), [Validação: success](https://github.com/franklinlem/certguard/actions/runs/38294848666) |
| [#7](https://github.com/franklinlem/certguard/pull/7) | Agendamento automático de verificações TLS e retenção do histórico | feature/auth-endpoints-checks | [Segurança: success](https://github.com/franklinlem/certguard/actions/runs/38296862985), [Validação: success](https://github.com/franklinlem/certguard/actions/runs/38296862965) |
| [#8](https://github.com/franklinlem/certguard/pull/8) | Dashboard global, detalhes TLS e gráfico acessível do histórico | feature/scheduler-retention | [Segurança: success](https://github.com/franklinlem/certguard/actions/runs/38327418886), [Validação: success](https://github.com/franklinlem/certguard/actions/runs/38327418880) |
| [#9](https://github.com/franklinlem/certguard/pull/9) | Motor de alertas com deduplicação, recuperação e histórico de eventos | feature/dashboard | [Segurança: success](https://github.com/franklinlem/certguard/actions/runs/38396100906), [Validação: success](https://github.com/franklinlem/certguard/actions/runs/38396100910) |
| [#10](https://github.com/franklinlem/certguard/pull/10) | Canais cifrados e entregas de alertas por SMTP e webhook | feature/alert-events | [Segurança: success](https://github.com/franklinlem/certguard/actions/runs/38399819552), [Validação: success](https://github.com/franklinlem/certguard/actions/runs/38399819572) |

## Dependências de integração

#6 → main; #7 → #6; #8 → #7; #9 → #8; #10 → #9. Integrar em ordem e atualizar a base dos seguintes para main, executando CI sobre cada integração. Não fazer squash de uma base mantendo PRs filhos sem revisar ancestralidade e diffs.

Os commits finais de #6–#10 apenas registram evidências, por isso as execuções acima se referem ao commit de implementação imediatamente anterior (ou anterior à atualização documental). SHAs atuais:

- #6: `ceba4d43c5aa1393c857c23f4e2afa96a4acd4ea`.
- #7: `d795acdf8ede28de21193347d3e2bed23db55734`.
- #8: `10ef6ea7c5bf536b98a530c8dad46f4e27817cb4`.
- #9: `8248130158dc3a6c851082db5c1ad4dab99816d4`.
- #10: `46c3cd8d9e8e14ee4eb496a05ba8186a704fe720`.

Dependabot #1–#5 testa sobre main, ainda na fundação. Isso não valida automaticamente as alterações junto de autenticação, scheduler e alertas. Avaliar majors de Node/Actions e reconstruir imagens no estado integrado antes de aceitar. Nenhum PR automático foi mesclado ou fechado nesta revisão.

Conclusão: sem impedimento identificado para continuar desenvolvimento. A ausência de conflitos não equivale a revisão humana, integração concluída ou homologação da VPS. O próximo PR de hardening parte de #10.
