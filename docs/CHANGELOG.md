# Registro de cambios

## 2026-11-02 · Icono de GitHub en la home

- La home incluye un botón-icono en el hero que enlaza al repositorio `github.com/nestorPons/klowna` (nueva pestaña, `rel="noopener"`).
- Nuevo icono relleno `github` en `internal/render/icons.go` y clase `.btn-square` (46 × 46 px, radio 6 px).
- `style.css` pasa a `v=12`.

Archivos:

```text
internal/render/icons.go
web/templates/pages/home.html
web/templates/base.html
web/static/css/style.css
docs/CHANGELOG.md
```

## 2026-11-02 · Despliegue a producción

- Backup previo: `backups/pre-deploy-20261102-080544.sql.gz` en el servidor.
- Sincronizado el código a `/root/projects/klowna.com` y reconstruida la imagen `klowna-app`.
- Migraciones aplicadas: `019_history_cash.sql`, `020_deposits.sql`, `021_user_avatar.sql`.
- Se añade `TURNSTILE_SECRET` al `.env` del servidor. SMTP sigue vacío en producción.
- Verificado: `/` 200, `/login` 200, panel 200 y sidebar sin «Leads».

## 2026-11-02 · Un solo compose y dominio klowna.com

- Eliminado `docker-compose.prod.yml`; local y producción usan el mismo `docker-compose.yml`.
- El compose publica la app en `127.0.0.1:${APP_PORT}` y se conecta a la red externa `traefik-net` (en local: `docker network create traefik-net`).
- Router de Traefik: `Host(`klowna.com`) || Host(`www.klowna.com`)` con TLS `myresolver`; certificado Let's Encrypt emitido para ambos dominios.
- Cloudflare sigue delante del origen; no hubo cambios de DNS.

```text
docker-compose.yml
README.md
docs/CHANGELOG.md
```

## 2026-11-02 · Leads fuera del menú del panel

- Se elimina el enlace **Leads** del lateral del panel (`/app/leads`).
- La ruta y la vista siguen existiendo: solo se oculta el acceso desde el menú.

Archivos:

```text
web/templates/partials/app/sidebar.html
docs/CHANGELOG.md
```

## 2026-10-30 · Iniciales de trabajadores en la agenda

- Cuando la agenda oculta los iconos, cada trabajador se identifica con las iniciales de su **nombre y primer apellido** (por ejemplo, Ana Beltrán → `AB`) en lugar de las dos primeras letras del nombre.

Archivos:

```text
internal/render/render.go
docs/CHANGELOG.md
```

## 2026-10-30 · Revertida la rejilla de 15 minutos de la agenda

- Se deshacen los cambios de los dos ajustes anteriores: la rejilla vuelve a mostrar una línea por hora con `var(--line)`.
- El eje de horas vuelve a mostrar solo las etiquetas de hora (sin `:15`, `:30`, `:45`).
- `app.css` vuelve a `v=32`.

Archivos:

```text
web/static/css/app.css
web/templates/partials/app/agenda.html
web/templates/app.html
docs/CHANGELOG.md
```

## 2026-10-30 · Modal de cita más ancho

- El formulario flotante de edición de citas y servicios de la agenda usa `width: min(1440px, 100%)`.
- `app.css` pasa a `v=32`.

Archivos:

```text
web/templates/partials/app/agenda.html
web/static/css/app.css
web/templates/app.html
docs/CHANGELOG.md
```

## 2026-10-30 · Autoguardado de filas en equipo y servicios

- Las filas editables de **Configuración → Equipo** y del catálogo de **Servicios** se guardan automáticamente al perder el foco de la fila, sin pulsar «Guardar».
- Componente reutilizable `autosave.js`: detecta cambios, espera a que el foco salga de la fila y envía el formulario. Evita guardados intermedios al tabular entre campos.
- Equipo responde por HTMX con el fragmento de la sección (`settings-equipo`) y muestra un aviso «Guardado»; servicios ya actualizaba su bloque por HTMX.
- Se eliminan los botones «Guardar» de las filas; siguen funcionando el envío con Enter y el fallback sin JavaScript.
- Corregido `UpdateService` para aceptar guardados sin cambios (RowsAffected 0).
- `app.css` pasa a `v=31`; nuevo `autosave.js` (`v=1`).

Archivos:

```text
internal/handler/app.go
internal/store/app.go
web/templates/app.html
web/templates/fragments/settings-equipo.html (nuevo)
web/templates/partials/app/settings-equipo.html
web/templates/partials/app/services.html
web/static/css/app.css
web/static/js/components/autosave.js (nuevo)
docs/CHANGELOG.md
```

## 2026-10-30 · Sidebar sin selección de texto

- El lateral `#app-sidebar` bloquea la selección de texto con `user-select: none` (incluido prefijo WebKit).
- `app.css` pasa a `v=30`.

Archivos:

```text
web/static/css/app.css
web/templates/app.html
docs/CHANGELOG.md
```

## 2026-10-30 · Panel de conectados y perfiles de usuario

- **Zona de conectados**: el lateral muestra los usuarios con sesión activa (gestor y su equipo), con avatar/icono, rol y número de sesiones.
- **Edición de perfil**: desde cada usuario conectado (y desde Configuración → Equipo) se abre un modal para cambiar nombre, email, color, icono y avatar.
- **Avatar**: subida de imagen con recorte a 160 × 160 px en el navegador; se guarda en la base de datos y se sirve por `/app/usuarios/{id}/avatar`.
- **Contraseña**: el gestor puede resetear la de su equipo; al editar el propio perfil se exige la contraseña actual. Mínimo 8 caracteres.
- **Permisos**: cada usuario solo puede editar su propio perfil; el gestor también puede editar a su equipo.
- Migración `021_user_avatar.sql`; `app.css` a `v=29` y nuevo `users.js`.

Archivos:

```text
internal/database/migrations/021_user_avatar.sql (nuevo)
internal/model/app.go
internal/store/app.go
internal/handler/app.go
internal/handler/handler.go
internal/handler/user_test.go (nuevo)
internal/render/data.go
web/templates/partials/app/sidebar.html
web/templates/partials/app/settings-equipo.html
web/templates/app.html
web/static/css/app.css
web/static/js/components/users.js (nuevo)
README.md
docs/CHANGELOG.md
```

## 2026-10-30 · Cierre de caja simplificado y datos de prueba

- **Cierre de caja**: se elimina el conteo por billetes y monedas. El efectivo contado se introduce en un único campo numérico y el descuadre se calcula en vivo contra el efectivo esperado.
- **Historial de cajas**: se retira el detalle del conteo por denominaciones; se mantienen fondo, esperado, contado y descuadre.
- **Datos de prueba**: `scripts/dev-seed.sql` incluye precios, servicios con señal, citas con señal pendiente y cobrada, historial de clientes y una caja abierta con señales, cobro de factura y gasto.
- `app.css` a `v=27` y `caja.js` a `v=2`.

Archivos:

```text
internal/handler/app.go
internal/render/data.go
internal/render/render.go
internal/render/app_render_test.go
web/templates/partials/app/caja.html
web/templates/app.html
web/static/css/app.css
web/static/js/components/caja.js
scripts/dev-seed.sql
README.md
docs/CHANGELOG.md
```

## 2026-10-30 · Señal de reserva y formas de pago

- **Formas de pago**: efectivo, tarjeta, transferencia bancaria, bizum y otros, disponibles en cobros, facturas, caja y señal de reserva.
- **Señal por servicio**: cada servicio puede exigir una señal de importe fijo. Se configura en el catálogo (alta y edición).
- **Señal en la cita**: al crear o editar una cita, si algún servicio exige señal aparece el bloque «Señal de reserva» con el importe pendiente, la forma de pago y la casilla «Señal cobrada». La cita congela el importe exigido (`deposit_due`) y acumula lo abonado (`deposit_paid`).
- **Registro en caja**: al marcar la señal como cobrada se crea un ingreso en la caja del día con concepto «Señal cita · cliente» y vínculo a la cita. Al cobrar después la factura se descuenta la señal ya ingresada para no duplicar el importe. Si se elimina el movimiento de caja de una señal, la cita vuelve a mostrarse como pendiente.
- **Aviso visual**: las citas con señal pendiente muestran un punto rojo animado en agenda (día, semana y mes), con prioridad sobre el punto azul de nota, y una etiqueta «Señal» en los listados de inicio y ficha de cliente.
- Corregido: guardar una cita sin cambios en sus campos ya no falla al registrar solo la señal.
- Migración `020_deposits.sql`; `app.css` a `v=26` y `agenda.js` a `v=19`.

Archivos:

```text
internal/database/migrations/020_deposits.sql (nuevo)
internal/model/app.go
internal/model/app_test.go
internal/store/app.go
internal/handler/app.go
internal/render/app_render_test.go
web/templates/partials/app/agenda.html
web/templates/partials/app/services.html
web/templates/partials/app/caja.html
web/templates/app/dashboard.html
web/templates/app/patient.html
web/templates/app.html
web/static/css/app.css
web/static/js/components/agenda.js
README.md
docs/CHANGELOG.md
```

## 2026-10-30 · Configuración dividida en subsecciones

- La configuración se separa en cinco páginas: **Apariencia**, **Facturación**, **Horario del centro**, **Equipo** y **Bloqueos y festivos**.
- `/app/configuracion` redirige a `/app/configuracion/apariencia`; cada subsección tiene URL propia.
- El menú lateral convierte «Configuración» en un desplegable (`<details>`) que se abre en la subsección activa.
- Cada página incluye pestañas de navegación rápida y conserva los avisos de guardado/error.
- Los formularios redirigen a su subsección correspondiente tras guardar.
- `app.css` pasa a `v=26`.

Archivos:

```text
internal/handler/app.go
internal/handler/handler.go
internal/render/data.go
internal/render/app_render_test.go
web/templates/app.html
web/templates/app/settings-apariencia.html (nuevo)
web/templates/app/settings-facturacion.html (nuevo)
web/templates/app/settings-horario.html (nuevo)
web/templates/app/settings-equipo.html (nuevo)
web/templates/app/settings-bloqueos.html (nuevo)
web/templates/partials/app/settings-shell.html (nuevo)
web/templates/partials/app/settings-apariencia.html (nuevo)
web/templates/partials/app/settings-facturacion.html (nuevo)
web/templates/partials/app/settings-horario.html (nuevo)
web/templates/partials/app/settings-equipo.html (nuevo)
web/templates/partials/app/settings-bloqueos.html (nuevo)
web/templates/partials/app/sidebar.html
web/templates/partials/app/invoice-detail.html
web/static/css/app.css
README.md
docs/CHANGELOG.md
```

## 2026-10-30 · Historial de clientes, caja diaria y estados de cita

- **Historial del cliente**: nueva sección en la ficha con entradas tipadas (nota, comentario, seguimiento, incidencia), autor y fecha. Se pueden añadir, editar y eliminar. Incluye tabla de servicios realizados (citas realizadas, facturadas o abonadas).
- **Caja diaria**: apertura con fondo inicial, cobros y gastos por forma de pago, conteo de billetes y monedas, arqueo con descuadre, cierre con notas e historial de cajas.
- **Estados de cita**: pendiente → confirmada → realizada → no asistida → facturada → abonada. Al elegir «abonada» se muestra la forma de pago y el total con IVA; al guardar se genera la factura, se cobra y el ingreso se registra en la caja del día.
- **Facturación**: las facturas guardan la forma de pago y el cobro crea un movimiento de caja. Crear factura desde una cita la deja en estado «facturada».
- **Precio congelado**: las líneas de cita guardan el precio del servicio al crearse (antes quedaba a 0 y las facturas salían sin importe).
- Migración `019_history_cash.sql`; `app.css` a `v=25`, `agenda.js` a `v=18` y nuevo `caja.js`.

Archivos:

```text
internal/database/migrations/019_history_cash.sql (nuevo)
internal/model/app.go
internal/model/app_test.go (nuevo)
internal/store/app.go
internal/handler/app.go
internal/handler/handler.go
internal/render/data.go
internal/render/render.go
internal/render/app_render_test.go
web/templates/app.html
web/templates/app/caja.html (nuevo)
web/templates/app/patient.html
web/templates/app/dashboard.html
web/templates/partials/app/caja.html (nuevo)
web/templates/partials/app/agenda.html
web/templates/partials/app/invoice-detail.html
web/templates/partials/app/invoice-row.html
web/templates/partials/app/sidebar.html
web/static/css/app.css
web/static/js/components/agenda.js
web/static/js/components/caja.js (nuevo)
README.md
docs/CHANGELOG.md
```

## 2026-10-30 · Contacto con SMTP y captcha Turnstile

- Cada lead del formulario se envía por SMTP al buzón configurado en `SMTP_TO` (mismo remitente que nestorpons.com), con `Reply-To` del visitante. Si el correo falla, el lead queda guardado y se registra el error.
- Captcha Cloudflare Turnstile con site key `0x4AAAAAAFsU9zRaPwnUXFLc` y validación servidor contra `siteverify`.
- Fail closed: sin `TURNSTILE_SECRET` el formulario rechaza envíos. `TURNSTILE_SKIP=true` solo para desarrollo.
- CSP ampliada para `challenges.cloudflare.com`; el widget se reinicia tras cada envío.
- Composes y `.env.example` con `SMTP_*` y `TURNSTILE_*`.
- Pendiente: rellenar `TURNSTILE_SECRET` (clave privada del par de la site key) en `.env`.

Archivos:

```text
internal/mailer/mailer.go (nuevo)
internal/mailer/mailer_test.go (nuevo)
internal/captcha/captcha.go (nuevo)
internal/captcha/captcha_test.go (nuevo)
internal/config/config.go
internal/handler/handler.go
internal/model/model.go
internal/render/data.go
web/templates/partials/lead-form.html
web/templates/base.html
web/static/js/main.js
web/static/css/style.css
docker-compose.yml
docker-compose.prod.yml
.env
.env.example
README.md
docs/CHANGELOG.md
```

## 2026-10-29 · Datos demo ampliados

- **Servicios**: 19 servicios demo con color e icono únicos (sin repetir).
- **Equipo**: colores propios distintos a los del catálogo (Ana azul, Carlos verde, Lucía rosa).
- **Comida**: bloques cerrados de lunes a viernes (Ana 14:00–15:00, Carlos 14:30–15:30, Lucía 15:00–16:00).
- **Festivos**: domingos marcados como festivo durante las próximas 8 semanas, además del cierre semanal.
- **Notas**: notas añadidas a 12 citas de ejemplo (cabinas, preparación, seguimiento).
- Seed idempotente: limpia bloques y festivos demo antes de insertar.

Archivos:

```text
scripts/dev-seed.sql
docs/CHANGELOG.md
```

## 2026-10-29 · Despliegue a producción

- Backup previo: `backups/pre-deploy-20261029-135154.sql.gz` en el servidor.
- Sincronizado el código a `/root/projects/klowna.com` y reconstruido `klowna-app`.
- Migración `018_schedule_block_series.sql` aplicada; healthcheck OK y panel de configuración verificado.

## 2026-10-29 · Bloques de empleado editables, horas 24 h y «todos los días»

- «Horas cerradas por empleado» pasa a una lista desplegable: cada bloque se puede editar (empleado, día, horas y motivo) o eliminar.
- Nueva opción «Todos los días» al crear o editar un bloque. Crea/actualiza el bloque para los 7 días mediante `series_id`; al editar un bloque individual con la opción marcada se agrupa y replica, y al editar uno de la serie se actualizan todos.
- Las horas de los formularios (horario del centro, bloques y citas) usan campos de texto en formato 24 h `HH:MM` con validación y normalización (`7:00` → `07:00`).
- El servidor acepta y normaliza horas sin cero inicial; las citas mantienen la vista actual al mostrar errores.
- `app.css` pasa a `v=23`, `agenda.js` a `v=17` y se añade `time24.js` (`v=1`).

Archivos:

```text
internal/database/migrations/018_schedule_block_series.sql (nuevo)
internal/model/app.go
internal/store/app.go
internal/handler/app.go
internal/handler/handler.go
web/templates/app/settings.html
web/templates/partials/app/agenda.html
web/templates/app.html
web/static/css/app.css
web/static/js/components/agenda.js
web/static/js/components/time24.js (nuevo)
docs/CHANGELOG.md
```

## 2026-10-29 · Horas cerradas por empleado en la agenda

- Los tramos cerrados del empleado bloquean la agenda en día y semana: se ocultan los `+` de las franjas afectadas y las celdas muestran el rayado de cerrado con su motivo.
- Los tramos del horario del centro también ocultan sus `+` en la vista de día.
- El arrastre rechaza soltar citas sobre tramos cerrados (atributo `data-closed`) y avisa al usuario.
- Validación en servidor al crear, editar y mover citas: responde con error si la cita se solapa con un bloque del empleado asignado. Los errores de creación conservan la vista actual.
- `agenda.js` pasa a `v=16`.

Archivos:

```text
internal/handler/app.go
internal/render/render.go
internal/render/data.go
web/templates/partials/app/agenda.html
web/static/js/components/agenda.js
web/templates/app.html
docs/CHANGELOG.md
```

## 2026-10-29 · Clic en las celdas horarias (+) de la agenda

- El listener global de «clic fuera» cerraba el panel «Nueva cita» en el mismo clic que lo abría desde una celda horaria.
- El cierre exterior ahora ignora los clics sobre `[data-add-slot]`; pulsar varias celdas actualiza la hora sin cerrar el panel.
- `agenda.js` pasa a `v=15`.

Archivos:

```text
web/static/js/components/agenda.js
web/templates/app.html
docs/CHANGELOG.md
```

## 2026-10-29 · Persistencia de los iconos de la agenda al cambiar de vista

- Al cambiar entre Día/Semana/Mes, HTMX reemplazaba `#agenda` después de `htmx:afterSwap` y el ajuste de atributos de HTMX eliminaba la clase `is-icons-hidden`, por lo que la preferencia de iconos se perdía.
- Los componentes se inicializan ahora en `htmx:afterSettle` (después del ajuste de HTMX), de modo que la vista, la fecha, el trabajador y los iconos se conservan al cambiar de vista o de sección.
- `app.js` pasa a `v=8`.

Archivos:

```text
web/static/js/app.js
web/templates/app.html
docs/CHANGELOG.md
```

## 2026-10-29 · Plan destacado a ancho completo en precios

- `/precios`: el plan «App gratuita» ocupa toda la primera fila de la rejilla.
- `style.css` pasa a `v=10`.

Archivos:

```text
web/static/css/style.css
web/templates/base.html
web/templates/auth.html
web/templates/app.html
docs/CHANGELOG.md
```

## 2026-10-29 · Persistencia de filtros al recargar

- Agenda: la vista, la fecha, el trabajador seleccionado y la visibilidad de iconos se conservan en la URL y en `localStorage`; al recargar o volver a `/app/agenda` se restaura el último estado.
- Facturación: el filtro de estado (`Todas`, `Pendientes`, `Pagadas`) se refleja en la URL con `hx-push-url`, por lo que se mantiene al recargar.
- Clientes: la búsqueda actualiza la URL con `hx-replace-url` (sin recargar ni perder el foco), y al recargar se conserva la consulta y la lista filtrada.
- `agenda.js` pasa a `v=14`.

Archivos:

```text
internal/handler/app.go
web/templates/partials/app/invoices.html
web/templates/partials/app/patients.html
web/templates/app.html
web/static/js/components/agenda.js
docs/CHANGELOG.md
```

## 2026-10-29 · Iconos sólidos en servicios y equipo

- Nuevo helper `solidIcon` con versiones rellenas de los iconos de servicios y empleados (`sparkles`, `star`, `heart`, `leaf`, `flame`, `sun`, `droplet`, `eye`, `smile`, `scissors`, `zap`, `stethoscope`, `clock`, `user`).
- Los iconos numéricos (`num-0`…`num-9`) se mantienen como dígitos rellenos.
- Aplicado en agenda (día y semana), listado de servicios y avatares del equipo.
- Se actualiza la captura de la landing con el nuevo estilo.

Archivos:

```text
internal/render/icons.go
internal/render/render.go
web/templates/partials/app/agenda.html
web/templates/partials/app/services.html
web/templates/app/settings.html
web/templates/pages/home.html
web/static/img/dashboard-agenda-day.png
docs/CHANGELOG.md
```

## 2026-10-29 · Captura del panel de agenda en la landing

- Nueva sección «La agenda en acción» justo después del hero, con una captura real del panel de agenda (día, profesionales, servicios con color e icono, notas).
- Imagen optimizada en `web/static/img/dashboard-agenda-day.png` (2880×1800, `loading="lazy"`).
- Nuevos estilos `.dashboard-showcase` y `.dashboard-figure`.
- `style.css` pasa a `v=9`.

Archivos:

```text
web/templates/pages/home.html
web/static/img/dashboard-agenda-day.png
web/static/css/style.css
web/templates/base.html
web/templates/app.html
web/templates/auth.html
docs/CHANGELOG.md
```

## 2026-10-29 · Búsqueda de clientes sin perder el foco

- El buscador de clientes solo actualiza la lista (`#patients-list`) y no reemplaza el campo, por lo que mantiene el foco y el cursor mientras se escribe.
- Nuevo fragmento `patients-list` y bloque reutilizable `patients-list-content`.
- Se añade el evento `search` al disparador (limpiar con Escape también refresca).

Archivos:

```text
internal/handler/app.go
web/templates/partials/app/patients.html
web/templates/fragments/patients-list.html
docs/CHANGELOG.md
```

## 2026-10-29 · Filtro por trabajador en la semana y señal de notas

- La vista semanal incorpora un selector para filtrar por trabajador (gestores) y un doble clic en la cabecera/icono de un trabajador aplica ese filtro.
- Con el filtro activo se muestra solo la columna del trabajador y las citas pasan a modo expandido: servicios con color, icono, nombre y duración, paciente, horario y notas.
- El filtro se conserva al navegar entre fechas, vistas y al crear/editar citas.
- Las tarjetas con notas muestran un punto azul sobre el borde superior derecho, sin texto, para no restar espacio a los servicios en citas cortas.
- Los iconos de servicio de la semana compacta se alinean a la izquierda.
- Al ocultar los iconos, las cabeceras semanales muestran las dos primeras letras del trabajador en mayúsculas.
- Nuevo icono `note`.
- La altura de la vista semanal pasa a 64px por hora (16px por bloque de 15 min) para que quepa al menos una línea de contenido.
- Los iconos dentro de botones `.btn` se limitan a 18×18 px (el icono de «Añadir festivo» se veía enorme).
- El estado de visibilidad de los iconos se guarda en `localStorage` (`klowna.agenda.icons`) y se aplica al cargar y al cambiar de vista.
- En la agenda, pulsar en cualquier punto de una celda horaria abre «Nueva cita» (antes solo el botón `+`).
- Se actualizan `app.css` a `v=22`, `style.css` a `v=8` y `agenda.js` a `v=13`.

Archivos:

```text
internal/handler/app.go
internal/render/data.go
internal/render/icons.go
internal/render/render.go
web/templates/partials/app/agenda.html
web/templates/app.html
web/templates/base.html
web/templates/auth.html
web/static/css/app.css
web/static/css/style.css
web/static/js/components/agenda.js
docs/CHANGELOG.md
```

## 2026-10-29 · CTA del hero hacia login y script de aplicación rápida

- El botón del hero de la home pasa de «Solicita una demo» (`#demo`) a «Iniciar sesión» (`/login`).
- Nuevo `scripts/apply-dev.sh`: compila el binario en local y lo inyecta en el contenedor `app` sin reconstruir la imagen Docker.

Archivos:

```text
web/templates/pages/home.html
scripts/apply-dev.sh
README.md
docs/CHANGELOG.md
```

## 2026-10-28 · Switch para mostrar u ocultar los iconos de la agenda

- Nuevo switch «Iconos» en los controles de la agenda (cabecera).
- Oculta o muestra los iconos de empleados y servicios en las vistas día, semana y mes.
- La preferencia se guarda en `localStorage` (`klowna.agenda.icons`) y se conserva al cambiar de vista.
- Se actualizan `app.css` a `v=15` y `agenda.js` a `v=11` en `app.html`.

Archivos:

```text
web/templates/partials/app/agenda.html
web/templates/app.html
web/static/css/app.css
web/static/js/components/agenda.js
docs/CHANGELOG.md
```

## 2026-10-28 · Panel «Nueva cita» al 80% del contenido

- El formulario flotante de la agenda pasa a ocupar el **80%** del ancho del contenido principal (referencia: `.agenda-head`).
- Se posiciona respecto a `.agenda-head` y arranca bajo toda la cabecera.
- Se actualiza `app.css` a `v=14` en `app.html` para invalidar la caché.

Archivos:

```text
web/templates/app.html
web/static/css/app.css
docs/CHANGELOG.md
```

## 2026-10-28 · Botón «Nueva cita» en la cabecera de la agenda

- El botón «Nueva cita» se mueve dentro de `.agenda-head` (junto a los controles de vista/fecha).
- El formulario se abre como **panel flotante** anclado bajo el botón, sin comprimir la cabecera ni empujar la agenda.
- En móvil el panel ocupa el ancho disponible con scroll interno.
- El panel se cierra con `Escape` o al hacer clic fuera; se elimina el `scrollIntoView` que ya no aplica.
- Se actualizan las versiones de `app.css` y `agenda.js` en `app.html` para invalidar la caché.

Archivos:

```text
web/templates/partials/app/agenda.html
web/templates/app.html
web/static/css/app.css
web/static/js/components/agenda.js
docs/CHANGELOG.md
```

## 2026-10-28 · Citas con varios servicios: un tramo por servicio

- Las citas con varios servicios se dividen en tramos proporcionales a la duración de cada servicio, mostrando en cada tramo el **color e icono** de su servicio (día: icono + nombre + duración; semana: icono).
- La tarjeta deja de usar el color del primer servicio y pasa a componerse de los tramos de cada servicio.
- Bajo los tramos se mantiene una línea compacta con paciente y horario.

Archivos:

```text
web/templates/partials/app/agenda.html
web/static/css/app.css
web/templates/app.html
web/templates/auth.html
internal/render/app_render_test.go
docs/CHANGELOG.md
```

## 2026-10-28 · Franjas de 15 minutos para crear citas y arrastrar

- Las celdas horarias de la agenda (día y semana) se dividen en **franjas de 15 minutos**; el botón «+» aparece en cada franja y crea la cita con esa hora exacta (p. ej. `08:15`).
- El arrastre de citas ya ajusta a **intervalos de 15 minutos**.
- Los campos de hora de creación/edición usan `step="900"` para seleccionar en tramos de 15 minutos.

Archivos:

```text
internal/render/render.go
web/templates/partials/app/agenda.html
web/static/css/app.css
web/templates/app.html
web/templates/auth.html
docs/CHANGELOG.md
```

## 2026-10-28 · Horario del centro, días cerrados, festivos y color de las cards

- «Horario de la clínica» pasa a llamarse **«Horario del centro»**.
- Cada día del horario tiene una casilla **«Cerrado»**; los días cerrados **no aparecen en la vista de semana**.
- Nueva sección **«Festivos»** (alta/baja por fecha y nombre). Los festivos aparecen como **cerrados y no editables** en día, semana y mes: sin botón «+», sin soltar citas y con fondo rayado.
- Las **cards de la vista de semana** usan el **color del servicio en toda la card** (se elimina el borde de color de la izquierda).
- Nueva migración `017_holidays.sql` (tabla `holidays`). Datos demo: domingo cerrado y un festivo de ejemplo.

Archivos:

```text
internal/database/migrations/017_holidays.sql (nuevo)
internal/model/app.go
internal/store/app.go
internal/handler/app.go
internal/handler/handler.go
internal/render/data.go
internal/render/render.go
internal/render/app_render_test.go
web/templates/app/settings.html
web/templates/partials/app/agenda.html
web/static/css/app.css
web/templates/app.html
web/templates/auth.html
scripts/dev-seed.sql
docs/CHANGELOG.md
```

## 2026-10-28 · Crear cita desde la celda horaria con "+" al pasar el ratón

- En las vistas **día** y **semana**, cada franja horaria muestra un **botón «+»** que solo aparece al pasar el ratón por la celda.
- Al pulsarlo se abre el desplegable **«Nueva cita»** con la **hora** de la celda (editable) y, en la semana, también el **día** y el **empleado** de esa columna.
- El botón se genera por franja horaria (`hourSlots`) solo para gestores.

Archivos:

```text
internal/render/data.go
internal/render/render.go
internal/render/app_render_test.go
web/templates/partials/app/agenda.html
web/static/css/app.css
web/static/js/components/agenda.js
web/templates/app.html
web/templates/auth.html
docs/CHANGELOG.md
```

## 2026-10-28 · Iconos de trabajadores, agenda por horas y estado de vista persistente

- **Trabajadores**: los empleados admiten **icono y color**, asignables desde Configuración (nueva sección «Equipo»); se muestran en la cabecera de la vista de día y en la semana. Guardar sin cambios ya no devuelve error.
- **Vista de semana**: rediseñada como un calendario de **horas en el eje Y** (a la izquierda) y **días en el eje X**, con cada día subdividido por empleado; las citas ocupan su franja horaria y muestran **icono y color** del servicio al ser columnas estrechas.
- **Persistencia de la vista**: la vista de la agenda (día/semana/mes) y su fecha se guardan en `localStorage`; al volver a entrar se restaura la última vista usada.
- **Iconos numéricos**: nuevos iconos `num-0`…`num-9` (dígito en SVG) disponibles para servicios y equipo; los trabajadores demo usan `num-1`, `num-2` y `num-3`.
- **Datos de prueba**: los servicios demo incluyen **color e icono** distintos y los trabajadores demo reciben icono y color.
- Nueva migración `016_worker_style.sql` (columnas `color` e `icon` en `users`).

Archivos:

```text
internal/database/migrations/016_worker_style.sql (nuevo)
internal/model/app.go
internal/store/app.go
internal/handler/app.go
internal/handler/handler.go
internal/render/data.go
internal/render/render.go
internal/render/app_render_test.go
web/templates/partials/app/agenda.html
web/templates/app/settings.html
web/templates/app.html
web/templates/auth.html
web/static/js/components/agenda.js
web/static/css/app.css
scripts/dev-seed.sql
docs/CHANGELOG.md
```

## 2026-10-28 · Refinado visual, servicios con color/icono y horarios cerrados

- **Layout**: se elimina el bloque de usuario del topbar (menú móvil pasa a botón flotante) y los paddings de `.app-content`; se quita el enlace duplicado de Configuración en el sidebar.
- **Servicios**: cada servicio admite **color e icono**, visibles en el catálogo y en las citas de la agenda.
- **Vista de semana**: se reordena como matriz **días × usuarios** (columnas con las siglas de cada empleado); las citas muestran solo el color y el icono del servicio.
- **Horario de la clínica**: nueva sección en Configuración para definir apertura/cierre por día; las horas cerradas se sombrean en la agenda.
- **Bloques cerrados por empleado**: se pueden añadir tramos de hora cerrados por empleado, que aparecen marcados como cerrados en la agenda de día.
- **Texto**: se sustituye «paciente» por «cliente» en el panel.
- **Estilo**: border-radius reducido a 6px en botones, tarjetas, mensajes y marcos (los avatares siguen siendo redondos) y mejora del foco/contraste de los inputs.

Archivos:

```text
internal/database/migrations/014_services_style.sql (nuevo)
internal/database/migrations/015_schedule.sql (nuevo)
internal/model/app.go
internal/store/app.go
internal/handler/app.go
internal/handler/handler.go
internal/render/data.go
internal/render/icons.go
internal/render/render.go
web/templates/app.html
web/templates/partials/app/topbar.html
web/templates/partials/app/sidebar.html
web/templates/partials/app/services.html
web/templates/partials/app/agenda.html
web/templates/partials/app/patients.html
web/templates/partials/app/invoices.html
web/templates/app/settings.html
web/templates/app/dashboard.html
web/templates/app/patient.html
web/templates/auth/login.html
web/static/js/components/agenda.js
web/static/css/theme.css
web/static/css/style.css
web/static/css/app.css
docs/CHANGELOG.md
```

## 2026-10-27 · Facturación: precio congelado, PDF, antiduplicado y edición

- El **precio de cada servicio se congela** en la línea de la cita; la factura usa el precio acordado en la cita aunque luego cambie el precio del servicio.
- **Factura única por cita**: índice único sobre `invoices.appointment_id` y detección de duplicados en el servidor (pulsar «Crear factura» varias veces no genera duplicados).
- **Descarga de la factura en PDF** (botón «Descargar PDF» en el detalle), generado con `gopdf` y fuente Carlito embebida.
- **Edición de facturas pendientes**: desde el detalle, botón «Editar» permite cambiar concepto, IVA y líneas (añadir/eliminar/modificar descripción, cantidad y precio); los totales se recalculan.
- Nueva migración `013_invoice_pdf_edit.sql`.

Archivos:

```text
internal/database/migrations/013_invoice_pdf_edit.sql (nuevo)
internal/pdf/invoice.go (nuevo)
internal/pdf/invoice_test.go (nuevo)
internal/pdf/fonts/Carlito-Regular.ttf (nuevo)
internal/pdf/fonts/Carlito-Bold.ttf (nuevo)
internal/store/app.go
internal/handler/app.go
internal/handler/handler.go
go.mod
go.sum
web/templates/partials/app/invoice-detail.html
web/templates/partials/app/invoice-edit.html (nuevo)
web/templates/fragments/invoice-edit.html (nuevo)
web/static/js/components/invoice.js
web/static/css/app.css
docs/CHANGELOG.md
```

## 2026-10-27 · Facturación desde la agenda

- Desde el modal de edición de una cita en `/app/agenda` se puede **crear la factura de la cita**, reutilizando paciente y servicios sin teclear datos de nuevo. Si la cita ya tiene factura, se reutiliza.
- La factura muestra el **desglose por servicio y de IVA** (base imponible, % de IVA y total). El IVA por defecto es 21% y es **configurable**.
- Nueva sección **Configuración** (`/app/configuracion`) con los **datos fiscales del emisor** (razón social, NIF/CIF, dirección, teléfono, email) y el IVA por defecto.
- Los **servicios** incorporan un campo de **precio** (sin IVA), editable al crear/editar el catálogo.
- **Visualización de facturas**: doble clic sobre una fila de la lista (o botón «Ver») abre la factura completa en un modal, con opción de imprimir y de marcarla como pagada.
- Nueva migración `012_billing.sql` (precios de servicios, líneas de factura, IVA y datos fiscales por gestor).

Archivos:

```text
internal/database/migrations/012_billing.sql (nuevo)
internal/model/app.go
internal/store/app.go
internal/handler/app.go
internal/handler/handler.go
internal/render/data.go
internal/render/app_render_test.go (nuevo)
web/templates/app.html
web/templates/app/settings.html
web/templates/app/invoice.html (nuevo)
web/templates/app/patient.html
web/templates/partials/app/sidebar.html
web/templates/partials/app/agenda.html
web/templates/partials/app/invoices.html
web/templates/partials/app/invoice-row.html
web/templates/partials/app/invoice-detail.html (nuevo)
web/templates/partials/app/services.html
web/templates/partials/app/settings.html (nuevo)
web/templates/fragments/invoice-detail.html (nuevo)
web/static/css/app.css
web/static/js/components/agenda.js
web/static/js/components/invoice.js (nuevo)
docs/CHANGELOG.md
```

## 2026-10-27 · Precio de despliegue (/precios)

- Nuevo plan **Despliegue**: 25 € pago único en servidor propio; desde 60 € en servidor del cliente con panel y desde 90 € en VPS con SSH/Docker.
- La nota de precios detalla las condiciones del despliegue.
- Nuevas FAQ: coste del despliegue y factores que influyen en el precio en servidor del cliente.
- Nueva migración `012_deployment_pricing.sql`.

Archivos:

```text
web/templates/pages/pricing.html
internal/database/migrations/012_deployment_pricing.sql (nuevo)
docs/CHANGELOG.md
```

## 2026-10-27 · Ancho del panel de citas

- El contenido de la app ocupa el 100% del ancho disponible hasta un máximo de 1440px (antes 1080px).
- Se actualiza la versión del CSS en `app.html` para invalidar la caché.

Archivos:

```text
web/static/css/app.css
web/templates/app.html
docs/CHANGELOG.md
```

## 2026-10-27 · Renombrado del proyecto a Klowna

- Marca, textos, correos demo, contraseña demo, cookie de sesión y módulo Go pasan de `AiTurnia` a `Klowna` / `klowna`.
- Dominio de correos demo: `@aiturnia.es` → `@klowna.com`.
- Base de datos, usuario, contraseña, imagen, contenedores y volumen de Docker renombrados a `klowna`.
- Nueva migración `010_rebrand_klowna.sql` actualiza contenido y usuarios en bases existentes.
- El usuario de prueba `prueba@klowna.com` conservaba la contraseña antigua (`clowww1234`); se restablece a `klowna1234`.
- La carpeta del proyecto pasa a `/home/admin/Proyectos/klowna.com`; se conservan los datos mediante copia y restauración de la base.

Archivos:

```text
.env
.env.example
README.md
go.mod
docker-compose.yml
cmd/server/main.go
internal/config/config.go
internal/database/migrations/002_seed.sql
internal/database/migrations/003_content.sql
internal/database/migrations/004_app.sql
internal/database/migrations/005_agenda.sql
internal/database/migrations/007_contact.sql
internal/database/migrations/010_rebrand_klowna.sql (nuevo)
internal/handler/app.go
internal/handler/auth.go
internal/handler/handler.go
internal/render/data.go
internal/render/render.go
internal/render/render_test.go
internal/store/app.go
internal/store/store.go
scripts/dev-seed.sql
scripts/seed-dev.sh
web/static/js/app.js
web/static/js/components/agenda.js
web/static/js/components/core.js
web/static/js/components/menu.js
web/templates/auth/login.html
docs/CHANGELOG.md
```

## 2026-10-27 · Renombrado del proyecto a AiTurnia

- Marca, textos, correos demo, contraseña demo, cookie de sesión y módulo Go pasan de `clowww` a `AiTurnia` / `aiturnia`.
- Base de datos, usuario, contraseña, imagen, contenedores y volumen de Docker renombrados a `aiturnia`.
- Nueva migración `009_rebrand.sql` actualiza contenido y usuarios en bases existentes.
- La carpeta del proyecto pasa a `/home/admin/Proyectos/AiTurnia`; se conservan los datos mediante copia y restauración de la base.

Archivos:

```text
.env
.env.example
README.md
go.mod
docker-compose.yml
cmd/server/main.go
internal/config/config.go
internal/database/migrations/002_seed.sql
internal/database/migrations/003_content.sql
internal/database/migrations/004_app.sql
internal/database/migrations/005_agenda.sql
internal/database/migrations/007_contact.sql
internal/database/migrations/009_rebrand.sql (nuevo)
internal/handler/app.go
internal/handler/auth.go
internal/handler/handler.go
internal/render/data.go
internal/render/render.go
internal/render/render_test.go
internal/store/app.go
internal/store/store.go
scripts/dev-seed.sql
scripts/seed-dev.sh
web/static/js/app.js
web/static/js/components/agenda.js
web/static/js/components/core.js
web/static/js/components/menu.js
web/templates/auth/login.html
docs/CHANGELOG.md
```

## 2026-10-27 · Eliminación del blog

- Se eliminan las rutas `/blog` y `/blog/{slug}` y sus handlers.
- Se borran las plantillas `pages/blog.html` y `pages/post.html`.
- Se retiran los enlaces a Blog del header y del footer.
- Se eliminan `Posts`/`Post` de `PageData` y las referencias en tests.

Archivos:

```text
web/templates/pages/blog.html (eliminado)
web/templates/pages/post.html (eliminado)
web/templates/partials/header.html
web/templates/partials/footer.html
internal/handler/handler.go
internal/render/data.go
internal/render/render_test.go
docs/CHANGELOG.md
```

## 2026-10-27 · Precios reales (/precios)

- La aplicación tal y como está se presenta como **gratis** (plan destacado "Gratis").
- Desarrollo a medida: **desde 45 €**.
- Alojamiento en servidores propios: **10 €/mes** con contratación anual.
- Servicio técnico: **5 €/mes** con contratación anual.
- Nueva migración `008_pricing.sql` reemplaza los planes y las FAQ antiguas.
- La rejilla de precios pasa a 4 columnas adaptables.

Archivos:

```text
web/templates/pages/pricing.html
web/static/css/style.css
internal/handler/handler.go
internal/database/migrations/008_pricing.sql (nuevo)
docs/CHANGELOG.md
```

## 2026-10-27 · Panel: área principal en claro

- La sidebar del panel mantiene el tema oscuro navy; topbar, contenido, tarjetas, tablas, formularios y agenda pasan a fondo claro con texto navy y acentos menta.
- Se fija `color-scheme: light` en el panel para que los `select`, inputs de fecha/hora y menús nativos se rendericen en claro.
- Botones, enlaces, badges de estado, chips de agenda y modales adaptados al nuevo contraste.

Archivos:

```text
web/static/css/app.css
docs/CHANGELOG.md
```

## 2026-10-27 · Menú Soluciones: enlace a Bot WhatsApp

- Se eliminan los enlaces generados desde `Solutions` en el desplegable "Soluciones".
- Se añade un único enlace "Bot WhatsApp" a `https://aibotia.com` (nueva pestaña).

Archivos:

```text
web/templates/partials/header.html
docs/CHANGELOG.md
```

## 2026-10-27 · Identidad nestorpons, icono de agenda y landing a medida

- Tema oscuro navy/menta (`#081320`, `#3be38b`) en web pública y panel, con tipografía Antonio/Google Sans Flex, botones píldora y tarjetas de 20px.
- Nuevo icono de marca de agenda de citas en header, sidebar, login y favicon.
- Landing reenfocada: hero de agenda electrónica, banda de ventajas, características, cómo funciona, sectores y nueva sección de software a medida con precio competitivo; se retiran novedades, testimonios y blog.
- Footer reescrito según la aplicación: enlaces útiles y contacto real; se eliminan oficinas y teléfonos ficticios.
- Contacto unificado con nestorpons.com: `nestorpons@gmail.com`, `+34 660 291 797`, WhatsApp, web, GitHub y LinkedIn.
- Migración `007_contact.sql` actualiza `site_settings` en bases existentes.

Archivos:

```text
web/static/css/style.css
web/static/css/app.css
web/templates/partials/header.html
web/templates/partials/app/sidebar.html
web/templates/auth/login.html
web/templates/pages/home.html
web/templates/pages/contact.html
web/templates/pages/pricing.html
web/templates/partials/footer.html
web/static/img/favicon.svg
internal/database/migrations/002_seed.sql
internal/database/migrations/007_contact.sql (nuevo)
docs/CHANGELOG.md
```

## 2026-10-27 · Agenda: servicios en las tarjetas

- Los servicios de cada cita se muestran como lista vertical (`ul`/`li`), sin viñetas ni decoración, uno por línea.
- Se usa la plantilla reutilizable `agenda-svc-list` en las vistas de día y semana.
- Si la tarjeta no tiene altura suficiente, los servicios se recortan con elipsis.

Archivos:

```text
web/templates/partials/app/agenda.html
web/static/css/app.css
docs/CHANGELOG.md
```

## 2026-10-27 · Agenda: duración visual de las citas

- Causa raíz: el CSP (`style-src 'self'`) bloqueaba los `style` inline de la rejilla y de las tarjetas, por lo que no se aplicaban `top`/`height`.
- La rejilla de día y los eventos dejan de usar `style` inline; pasan a `data-cols`, `data-top` y `data-height`.
- `agenda.js` aplica la posición/altura y las columnas vía CSSOM, también tras cada swap de htmx (edición, creación, arrastre o cambio de vista).
- Altura mínima de tarjeta reducida a 15 min para que la duración real (15/30/45…) se distinga visualmente.
- Al añadir, quitar o cambiar servicios, el servidor recalcula `ends_at` y la tarjeta se redimensiona.

Archivos:

```text
web/templates/partials/app/agenda.html
web/static/js/components/agenda.js
internal/render/render.go
docs/CHANGELOG.md
```

## 2026-10-26 · Nueva identidad visual pública (aibotia.com)

- `style.css` adopta la paleta de aibotia.com: verde `#16a34a`, acentos `#22c55e`, grises Tailwind (`#f3f4f6`, `#e5e7eb`, `#1f2937`, `#111827`).
- Tipografía de sistema, botones y tarjetas con radios compactos (`rounded-lg`), sombras suaves y footer oscuro con enlaces verde claro.
- Heroes, bandas de stats, demo y CTA pasan de morado a verde oscuro/gris; se mantienen clases y plantillas.
- El panel SaaS (`app.css`), el login y el sidebar adoptan la misma paleta verde/gris.

Archivos:

```text
web/static/css/style.css
web/static/css/app.css
web/templates/base.html
web/templates/app.html
web/templates/auth.html
web/templates/auth/login.html
web/templates/partials/header.html
web/templates/partials/footer.html
web/templates/partials/app/sidebar.html
docs/CHANGELOG.md
```

## 2026-10-26 · Catálogo de servicios con duración

- Nueva tabla `services(id, user_id, name, duration_minutes, sort_order, created_at)` y página **Servicios** (`/app/servicios`, solo gestor) con alta, edición y borrado.
- `appointment_items` ahora referencia `service_id` (FK, `ON DELETE SET NULL`) y guarda `duration_minutes`; se elimina la columna de texto `service`.
- La duración de una cita se calcula como la **suma de las duraciones** de los servicios seleccionados; el formulario y el modal muestran un selector con casillas y el total en vivo.
- `POST /app/appointments` y `POST /app/appointments/{id}` aceptan IDs de servicio (`service_id` / `services`) y calculan `ends_at` en servidor.
- Migración `006_services.sql`: crea el catálogo, migra las líneas existentes (30 min por defecto) y refactoriza las líneas de cita.
- `dev-seed.sql` crea 19 servicios demo y asigna duraciones coherentes a las 18 citas.

Archivos:

```text
internal/database/migrations/006_services.sql (nuevo)
internal/model/app.go
internal/store/app.go
internal/handler/app.go
internal/handler/handler.go
internal/render/data.go
internal/render/render.go
web/templates/partials/app/services.html (nuevo)
web/templates/fragments/services.html (nuevo)
web/templates/app/services.html (nuevo)
web/templates/partials/app/agenda.html
web/templates/partials/app/sidebar.html
web/templates/app/dashboard.html
web/templates/app/patient.html
web/templates/app.html
web/static/css/app.css
web/static/js/components/agenda.js
scripts/dev-seed.sql
README.md
docs/CHANGELOG.md
```

## 2026-10-26 · Ajustes de agenda (vista de día y edición)

- Corregida la vista de día: rejilla de horas × empleados sin depender de `var()` en `repeat()` (compatibilidad), eje horario alineado (12 tramos) y eventos con altura mínima legible.
- Edición por **doble clic** sobre una cita (antes era clic simple); se añade aviso `title="Doble clic para editar"`.

Archivos:

```text
internal/handler/app.go
internal/render/render.go
web/templates/partials/app/agenda.html
web/static/css/app.css
web/static/js/components/agenda.js
docs/CHANGELOG.md
```

## 2026-10-26 · Agenda por empleado (día/semana/mes)

- Agenda con tres vistas: **día** (tabla horas × empleados), **semana** y **mes**, responsive con scroll horizontal en pantallas pequeñas.
- Asignación de citas a empleados: `demo` organiza y asigna; los empleados inician sesión y ven solo sus citas (solo lectura).
- Edición por arrastre (HTML5 DnD) y modal: cambiar empleado, hora, duración, estado y servicios.
- Servicios múltiples por cita (`appointment_items`); al eliminar el último servicio se elimina la cita.
- Roles `manager`/`staff` con jerarquía `manager_id`; rutas de gestión restringidas al gestor.
- Tres empleados demo: `ana@aiturnia.es`, `carlos@aiturnia.es`, `lucia@aiturnia.es` (contraseña `aiturnia1234`).
- Nuevos endpoints JSON: `POST /app/appointments/{id}`, `.../move` y `DELETE /app/appointments/{id}`.
- Migración `005_agenda.sql` y `dev-seed.sql` actualizado al nuevo esquema.

Archivos:

```text
internal/database/migrations/005_agenda.sql (nuevo)
internal/model/app.go
internal/store/app.go
internal/handler/app.go
internal/handler/handler.go
internal/render/data.go
internal/render/render.go
internal/render/icons.go
web/templates/partials/app/agenda.html
web/templates/partials/app/sidebar.html
web/templates/app/dashboard.html
web/templates/app/patient.html
web/templates/app.html
web/static/css/app.css
web/static/js/components/agenda.js (nuevo)
scripts/dev-seed.sql
README.md
docs/CHANGELOG.md
```

## 2026-10-26 · Datos de prueba para desarrollo

- Añadido `scripts/dev-seed.sql` con datos de prueba idempotentes: 16 pacientes, 18 citas, 12 facturas y 12 leads para el usuario demo.
- Añadido `scripts/seed-dev.sh` para aplicar el seed contra la base de Docker Compose.
- Documentado el seed en el README (sección «Datos»).

Archivos:

```text
scripts/dev-seed.sql (nuevo)
scripts/seed-dev.sh (nuevo)
README.md
docs/CHANGELOG.md
```

## 2026-10-26 · Aplicación SaaS de gestión

- Añadido el panel de gestión (agenda, pacientes y facturación) sobre la landing existente.
- Autenticación por sesión: login/logout, cookies HttpOnly y contraseñas con bcrypt (`golang.org/x/crypto`).
- Nuevas tablas `users`, `sessions`, `patients`, `appointments` e `invoices` con usuario y datos demo (migración `004_app.sql`).
- Vistas con HTMX (servido localmente) y JavaScript vanilla organizado en componentes reutilizables.
- Renderer multi-layout: marketing (`base`), login (`auth`), panel (`app`) y fragmentos HTMX.
- Rutas `/login`, `/app`, `/app/agenda`, `/app/pacientes`, `/app/facturacion` y `/app/leads`.
- Estilos propios del panel en `web/static/css/app.css`.

Archivos:

```text
internal/database/migrations/004_app.sql (nuevo)
internal/model/app.go (nuevo)
internal/store/app.go (nuevo)
internal/handler/auth.go (nuevo)
internal/handler/app.go (nuevo)
internal/handler/handler.go
internal/render/render.go
internal/render/data.go
internal/render/icons.go
web/templates/auth.html (nuevo)
web/templates/auth/login.html (nuevo)
web/templates/app.html (nuevo)
web/templates/app/{dashboard,agenda,patients,patient,invoices,leads}.html (nuevos)
web/templates/partials/app/{sidebar,topbar,agenda,patients,invoices,invoice-row}.html (nuevos)
web/templates/fragments/{agenda,patients,invoices,invoice-row}.html (nuevos)
web/static/css/app.css (nuevo)
web/static/js/vendor/htmx.min.js (nuevo)
web/static/js/components/{core,menu}.js (nuevos)
web/static/js/app.js (nuevo)
README.md
docs/CHANGELOG.md
```

## 2026-10-26 · Contenido real de flowww.es

- Sustituidos los planes inventados por las licencias reales Boss (59 €), Star (119 €) y Ultimate (249 €) con sus características y precio anual.
- Reemplazadas las preguntas frecuentes por las reales de la página de FAQ.
- Añadida la página `/preguntas-frecuentes` (plantilla `faq.html`, handler `FAQ` y ruta).
- Añadidos al home los logos de clientes reales (Clínicas Esquivel, 123Mobilité, Tacto, Be-beauty, Liverpool, Sundara, Epilpoint, etc.).
- Actualizado el footer con las oficinas, teléfonos y enlaces legales reales.
- Header: "Recursos" pasa a desplegable con Blog y Preguntas frecuentes.
- README: nueva ruta en la tabla de rutas.

Archivos modificados:

```text
internal/database/migrations/003_content.sql (nuevo)
web/templates/pages/faq.html (nuevo)
internal/handler/handler.go
web/templates/partials/header.html
web/templates/partials/footer.html
web/templates/pages/home.html
web/templates/pages/pricing.html
README.md
docs/CHANGELOG.md
```

## 2026-10-23 · Versión inicial

- Creado clon Go del sitio público con landing, precios, blog, sectores, especialidades y contacto.
- Añadida MariaDB con migraciones embebidas y datos semilla.
- Añadida captación de leads (JSON y formulario) y panel interno con Basic Auth.
- Añadido Dockerfile multi-stage y `docker-compose.yml` con healthchecks.
- Añadido diseño responsive propio con paleta de marca y mockups CSS.
- Verificado con `go test ./...`, `go vet ./...` y stack Docker en local.

Archivos creados:

```text
go.mod, go.sum, Dockerfile, docker-compose.yml, .env.example, .gitignore, .dockerignore
cmd/server/main.go
internal/config/config.go
internal/database/database.go
internal/database/migrations/001_schema.sql
internal/database/migrations/002_seed.sql
internal/store/store.go
internal/model/model.go
internal/render/render.go
internal/render/render_test.go
internal/render/data.go
internal/render/icons.go
internal/handler/handler.go
web/embed.go
web/templates/base.html
web/templates/partials/{header,footer,lead-form}.html
web/templates/pages/{home,pricing,blog,post,contact,sector,specialty,leads,notfound}.html
web/static/css/style.css
web/static/js/main.js
web/static/img/favicon.svg
README.md
docs/CHANGELOG.md
```
