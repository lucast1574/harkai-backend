# Municipal contact requests

`POST /v1/municipal-inquiries` accepts contact requests without account creation.
Required fields: requestId (UUID), municipality, name, role, email, message and
consent=true. Phone is optional. The response contains only an opaque receipt ID
and status; contact details are never returned by the public endpoint.

Requests become support records with kind=municipal and an atomically persisted
municipal_inquiry email job. UUID retries are idempotent. Per-IP, per-address and
global limits bound notification abuse. Admins can list them through
`GET /v1/admin/support/tickets?kind=municipal` and follow up using the existing
versioned ticket lifecycle. Closing a municipal request does not send the
technical-support resolution template to the contact.

Set `MUNICIPAL_CONTACT_EMAIL` to the internal recipient. Together with a valid
`RESEND_API_KEY` and `MAIL_FROM`, this enables the endpoint at startup. The address
is server configuration and must not be exposed in public bundles. Missing mail
configuration returns 503 instead of silently accepting unnotified requests.

The email worker reuses persisted leases and provider idempotency keys. Delivery
failures retry with backoff within the existing 23-hour outbox window. Requests
remain available in the admin inbox even if notification delivery expires.
The escaped HTML and plain-text templates live in internal/mail/templates.

Public form: https://panel.harkai.lat/municipalidades
Admin inbox: https://panel.harkai.lat/dashboard/admin?section=municipal

Municipal access is agreed directly; submitting a request does not create an
institutional role, grant access or charge a payment method. The map-listing
Stripe product remains separate.
