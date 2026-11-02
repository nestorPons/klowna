# Klowna

Klowna es una aplicación de agenda electrónica y gestión de citas para clínicas y centros, construida en Go y MariaDB. Incluye la web pública, captación de leads y un panel de gestión con agenda, pacientes, servicios y facturación. El texto base se inspira en flowww.es; el proyecto no está afiliado a flowww.

## Stack

- Go 1.24+ · `net/http` + `html/template`
- MariaDB 11
- Docker Compose
- CSS/JS propios (Cloudflare Turnstile es el único script externo)
- SMTP para avisos de contacto

## Arranque

```bash
cp .env.example .env
docker network create traefik-net   # solo la primera vez
docker compose up -d --build
```

- Web: http://localhost:8090 (`APP_PORT`)
- Panel de leads: http://localhost:8090/admin/leads (`ADMIN_USER` / `ADMIN_PASSWORD`)
- Healthcheck: http://localhost:8090/healthz
- MariaDB: `localhost:3308` (`DB_HOST_PORT`)

El compose publica la app solo en `127.0.0.1` y la conecta a la red externa `traefik-net`; en producción Traefik sirve el dominio y el puerto publicado queda como acceso directo local.

## Producción

- Dominio: **https://klowna.com** (también `www.klowna.com`) vía Cloudflare → Traefik.
- Un único `docker-compose.yml` para local y servidor.
- En el servidor (`/root/projects/klowna.com`):

```bash
docker compose up -d --build
docker compose ps
```

- Router de Traefik y TLS (LE, `myresolver`) definidos en los labels del servicio `app`. La red `traefik-net` debe existir en el host.

## Aplicar cambios

Con los contenedores en marcha, evita reconstruir la imagen: el builder clásico recompila todas las dependencias Go en cada build (~1 min).

```bash
scripts/apply-dev.sh
```

- Compila en local usando la caché de Go (~3 s).
- Copia el binario al contenedor `app` y lo reinicia.
- Sirve para cambios en Go, `.html`, CSS y JS (van embebidos en el binario).

Usa `docker compose up -d --build app` solo si cambian dependencias (`go.mod`/`go.sum`) o el `Dockerfile`.

## Aplicación (SaaS)

Panel de gestión incluido, con autenticación por sesión. Acceso: http://localhost:8080/login

| Cuenta | Email | Contraseña |
| --- | --- | --- |
| Demo (gestor) | `demo@klowna.com` | `klowna1234` |
| Empleada | `ana@klowna.com` | `klowna1234` |
| Empleado | `carlos@klowna.com` | `klowna1234` |
| Empleada | `lucia@klowna.com` | `klowna1234` |

El panel cubre la operativa diaria de una clínica:

- **Inicio** — métricas y citas del día.
- **Agenda** — vistas de día, semana y mes. En la vista de día, una tabla de horas × empleados permite organizar las citas arrastrándolas, cambiar el empleado asignado, los servicios (múltiples por cita), la hora o el estado, y eliminar la cita. La duración se calcula sumando la duración de los servicios elegidos. Estados: pendiente, confirmada, realizada, no asistida, facturada y abonada. Al marcar «abonada» se elige la forma de pago, se genera/cobra la factura y el ingreso queda registrado en caja. Las citas con señal pendiente muestran un punto rojo (con prioridad sobre el punto azul de nota). Los empleados inician sesión y ven únicamente sus propias citas.
- **Pacientes** — listado con búsqueda, alta y ficha con historial editable (notas, comentarios, seguimientos e incidencias), servicios realizados, citas, facturas y datos de contacto.
- **Servicios** — catálogo por clínica con nombre, duración, precio y señal de reserva opcional (importe fijo por servicio); se usa para componer las citas.
- **Facturación** — listado por estado, alta de facturas y cobro con forma de pago (efectivo, tarjeta, transferencia bancaria, bizum u otros).
- **Caja** — apertura diaria, cobros, señales de reserva y gastos por forma de pago, cierre con efectivo contado en un único campo, arqueo con descuadre e historial de cajas.
- **Conectados** — el lateral muestra quién tiene sesión activa y permite editar nombre, email, color, icono, avatar y contraseña.
- **Leads** — solicitudes recibidas desde la web.

Las vistas usan HTMX (servido localmente) para actualizaciones parciales y JavaScript vanilla organizado en componentes reutilizables (`web/static/js/components/`). Las filas editables de equipo y servicios guardan automáticamente al perder el foco.

## Rutas

| Ruta | Descripción |
| --- | --- |
| `/` | Landing |
| `/precios` | Planes y FAQ |
| `/blog`, `/blog/{slug}` | Listado y artículo |
| `/preguntas-frecuentes` | FAQ |
| `/contacto` | Formulario y datos de contacto |
| `/{sector}` | Sector (medicina-estetica, belleza-bienestar, salud) |
| `/{sector}/{especialidad}` | Especialidad |
| `POST /api/leads` | Alta de lead (JSON, requiere `cf-turnstile-response` si hay secret) |
| `POST /contacto` | Alta de lead (formulario con Turnstile) |
| `/login` | Acceso al panel (POST `/logout` para salir) |
| `/app` | Dashboard del panel |
| `/app/agenda` | Agenda diaria |
| `/app/pacientes`, `/app/pacientes/{id}` | Pacientes y su historial |
| `/app/servicios` | Catálogo de servicios (gestor) |
| `/app/facturacion` | Facturación y cobros |
| `/app/caja` | Caja diaria, arqueo e historial |
| `/app/usuarios/{id}` | Perfil del usuario conectado (datos, avatar y contraseña) |
| `/app/configuracion/{seccion}` | Configuración: apariencia, facturación, horario, equipo y bloqueos |
| `/app/leads` | Leads dentro del panel |
| `/admin/leads` | Panel interno con Basic Auth |
| `/healthz` | Estado de la app y la base de datos |

## Estructura

```text
cmd/server/                 arranque HTTP
internal/config/            variables de entorno y DSN
internal/database/          conexión + migraciones embebidas
internal/store/             consultas SQL (web y app)
internal/render/            plantillas, iconos y helpers
internal/handler/           rutas, validación, auth y middleware
internal/model/             modelos
web/templates/              layouts, páginas, parciales y fragmentos HTMX
web/static/                 CSS, JS (componentes + htmx) y favicon
docs/                       registro de cambios
```

## Datos

- Esquema y contenido inicial: `internal/database/migrations/*.sql`.
- Se aplican automáticamente al arrancar y se registran en `schema_migrations`.
- Tablas: `sectors`, `specialties`, `solutions`, `challenges`, `stats`, `testimonials`, `posts`, `plans`, `faqs`, `leads`, `site_settings`.
- Tablas del panel: `users` (con `role`, `manager_id`, `color`, `icon` y `avatar`), `sessions`, `patients`, `patient_history` (notas y comentarios del cliente), `services` (catálogo con `duration_minutes`, `price`, `deposit_required` y `deposit_amount`), `appointments` (con `assignee_id`, estado y `deposit_due`/`deposit_paid`), `appointment_items` (servicios por cita con `service_id`, `duration_minutes` y `price`), `invoices` (con `payment_method`), `cash_sessions` y `cash_movements`.
- Para reiniciar la base: `docker compose down -v && docker compose up -d --build`.
- Los leads del panel son las solicitudes enviadas desde la web. Si `SMTP_*` está configurado, cada lead se envía por correo a `SMTP_TO` con `Reply-To` del visitante.
- Datos de prueba para desarrollo: `scripts/seed-dev.sh` amplía el contenido demo del usuario `demo@klowna.com` (pacientes, citas, facturas) y añade leads de ejemplo. Es idempotente.

## Contacto y captcha

- Formulario: `web/templates/partials/lead-form.html`; `main.js` lo envía en JSON a `/api/leads`.
- Captcha: widget Turnstile con `TURNSTILE_SITE_KEY`; validación en servidor en `internal/captcha`.
- Sin `TURNSTILE_SECRET` el formulario rechaza envíos (fail closed). `TURNSTILE_SKIP=true` solo en desarrollo.
- Correo: `internal/mailer` envía el aviso por SMTP después de guardar el lead; si el correo falla, el lead queda guardado.

## Variables

| Variable | Defecto | Uso |
| --- | --- | --- |
| `APP_ADDR` | `:8080` | Escucha interna de Go |
| `APP_PORT` | `8090` | Puerto publicado de la web (solo `127.0.0.1`) |
| `DB_HOST` / `DB_PORT` | `127.0.0.1` / `3306` | Conexión a MariaDB |
| `DB_NAME` / `DB_USER` / `DB_PASSWORD` | `klowna` / `klowna` / `klownapass` | Base de datos |
| `DB_HOST_PORT` | `3307` | Puerto publicado de MariaDB |
| `DB_ROOT_PASSWORD` | `rootpass` | Root de MariaDB |
| `ADMIN_USER` / `ADMIN_PASSWORD` | `admin` / `admin` | Basic Auth del panel |
| `SMTP_HOST` / `SMTP_PORT` | — / `587` | Servidor SMTP de avisos |
| `SMTP_USER` / `SMTP_PASS` | — | Autenticación SMTP |
| `SMTP_FROM` / `SMTP_TO` | `SMTP_USER` / — | Remitente y destino de los avisos |
| `TURNSTILE_SITE_KEY` | `0x4AAAAAAFsU9zRaPwnUXFLc` | Clave pública del captcha |
| `TURNSTILE_SECRET` | — | Clave privada de Turnstile |
| `TURNSTILE_SKIP` | `false` | Desactiva el captcha (solo desarrollo) |

Cambia las credenciales por defecto antes de exponer el servicio fuera de tu equipo.

## Desarrollo sin Docker

```bash
# MariaDB local escuchando en 127.0.0.1:3306 y base/usuario creados
DB_HOST=127.0.0.1 DB_PORT=3306 DB_USER=klowna DB_PASSWORD=klownapass DB_NAME=klowna go run ./cmd/server
```

## Pruebas

```bash
go test ./...
go vet ./...
```

## Alcance

- Web pública, captación de leads y panel de gestión (agenda, pacientes, facturación y caja) funcionales.
- Panel con autenticación por sesión (email + contraseña con bcrypt) y sesiones en base de datos.
- Facturación con IVA, PDF, formas de pago y registro automático de cobros en caja.
- Sin pagos reales, historia clínica detallada ni multi-sede.
- Contenido y marca de Klowna; el texto base se inspira en flowww.es.
