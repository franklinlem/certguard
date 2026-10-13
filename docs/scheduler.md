# Operação do agendador

O comando `serve` inicia os workers junto da API e encerra todos ao receber SIGTERM/SIGINT. Sem habilitação, mantém somente verificações manuais. Não há novo serviço nem deploy automático.

| Variável | Padrão | Limites |
|---|---|---|
| SCHEDULER_ENABLED | false | booleano |
| CHECK_WORKERS | 10 | 1–25 por instância |
| CHECK_TIMEOUT | 10s | 1s–30s; inspeções automáticas |
| RETENTION_ENABLED | false | booleano, independente do scheduler |
| RETENTION_DAYS | 180 | 30–36500 |
| RETENTION_BATCH | 500 | 1–1000 por minuto/instância |

Configuração inválida impede a inicialização da API. O pool de tarefas reserva capacidade máxima de `CHECK_WORKERS + 2` conexões, separada do pool da API. Avalie conexões, memória e CPU antes de mudar workers ou réplicas.

Cadastre apenas destinos autorizados. Habilite `SCHEDULER_ENABLED=true` no ambiente escolhido e reinicie somente o serviço CertGuard pelo procedimento desse ambiente. Primeiro valide homologação. A interface exibirá “Agendamento habilitado”; use Atualizar para obter as datas atuais. Essa indicação representa configuração, não um heartbeat. Atrasos podem ocorrer por capacidade, indisponibilidade do banco ou reinício. Monitore logs e idade do último resultado.

Endpoints criados, reativados ou com destino/intervalo alterado recebem previsão nos próximos 30s; o cooldown de 30s desde o último check continua valendo. Pausados/arquivados não são selecionados. Um check já em andamento termina antes de a pausa/arquivamento adquirir o lock; alterações podem retornar conflito. Nome e limiares não alteram a previsão. Após conclusão, a próxima data usa intervalo + até 10% de dispersão.

Sem catch-up: após indisponibilidade, cada endpoint vencido recebe uma inspeção e uma nova agenda. Erros de TLS/DNS/SSRF são registrados e seguem a cadência normal. Erros de banco não avançam a agenda e geram log genérico; retries esperam 5s. Logs não incluem credenciais ou URL do banco.

Retenção remove dados definitivamente e fica desabilitada. Antes de habilitar num ambiente real, valide backup/restauração, revise idade/lote e obtenha autorização quando houver risco de perda de dados. A migração não apaga checks. Cada minuto processa no máximo um lote de antigos; o último resultado por endpoint permanece mesmo que ultrapasse a janela. Grandes históricos podem exigir vários ciclos.

Testes PostgreSQL usam schema descartável e inspeção injetada; não contornam a proteção SSRF. E2E usa ambas as opções desabilitadas, sem inspeções externas. Ainda não há alertas nem painel agregado.
