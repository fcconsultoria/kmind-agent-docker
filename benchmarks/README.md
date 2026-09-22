# Benchmark e aceite

## Objetivo

Medir CPU, RSS, goroutines, buffer, chamadas Docker Engine e dados descartados
nos cenarios de 1, 10, 50 e 100 containers. Execute em host Linux dedicado;
nunca em ambiente de cliente.

## Execucao reproduzivel

```bash
docker compose -f benchmarks/docker-compose.benchmark.yaml up -d --scale workload=50
docker stats --no-stream kmind-agent-docker
docker exec kmind-agent-docker /kmind-agent-docker --config /etc/kmind-agent/config.yaml
docker compose -f benchmarks/docker-compose.benchmark.yaml down -v
```

Repita cada cenario por 30 minutos, descartando os primeiros 5 minutos. Registre
RSS/CPU com `docker stats`, goroutines e buffer pelas metricas internas OTLP, e
chamadas Docker pelo contador interno do agente. Para indisponibilidade do Kmind,
aponte `KMIND_ENDPOINT` a um endpoint que retorne timeout; o RSS e buffer devem
permanecer dentro dos limites configurados.

## Matriz obrigatoria

| Cenario | Containers | Gateway | Logs |
| --- | ---: | --- | --- |
| Basico | 1 | saudavel | baixo |
| Medio | 10 | saudavel | medio |
| Carga | 50 | saudavel | alto |
| Limite | 100 | saudavel | alto |
| Falha | 10 | indisponivel | medio |

## Criterios

- CPU media <= 0,5% e RSS <= 30 MB devem ser medidos, nunca presumidos.
- Buffer deve ficar entre 0 e 5 MB, sem escrita de telemetria em disco.
- Logs devem ser descartados antes de metricas quando houver pressao.
- Nenhuma operacao mutavel da Docker API pode aparecer no trace de chamadas.

## Estado atual

Este protocolo esta pronto, mas o aceite esta **bloqueado** ate que coleta de
metricas, stream de eventos e stream de logs sejam conectados ao loop principal.
