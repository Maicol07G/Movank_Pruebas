# Decisiones técnicas

## Alcance
Full Stack, porque permite demostrar tanto experiencia web como persistencia y APIs.

## PostgreSQL como fuente de verdad
Las ventas no dependen de la cache. La cache puede perderse y reconstruirse desde PostgreSQL.

## Idempotencia
El cliente envía `Idempotency-Key`. La clave es única por tenant y apunta a la venta creada. Si llega el mismo request nuevamente, se devuelve la misma venta.

## UNKNOWN
Un timeout del proveedor no significa rechazo. La respuesta puede haberse perdido aunque el cobro haya ocurrido. Por eso se conserva `UNKNOWN` como estado auditable y no se crea otro cobro por reintentar la misma operación.

## Outbox
El registro de la venta y su evento se guardan en la misma transacción. El worker puede publicar/procesar después del commit sin depender de que el proceso HTTP siga vivo.

## Simplificaciones
Esta base mantiene algunas piezas deliberadamente pequeñas para que sean fáciles de estudiar. Antes de la entrega final hay que completar y probar el endpoint de pago, worker, SSE, autenticación real y el flujo offline completo.
