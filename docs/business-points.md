# Business points and permanent publications

Incident markers last three hours, including older pet/place records. MongoDB retains the incident, its moderated photo and comments; `/v1/publications` lists non-hidden records for the public website without authentication or geographic restrictions. Hiding a report removes its public detail, list and photo access.

Business points are separate records. Owners create/edit through `/v1/me/businesses`, admins approve/reject through `/v1/admin/businesses/{id}/moderation`. Public queries, photo access and proximity notices require enabled + approved + active access. During the free launch, the server grants payment_mode=free when no billing adapter is configured. No card or checkout is required. One-time paid records still require a valid paid_at; legacy subscriptions require future paid_until. Edits reset review. Unpaid legacy drafts with no payment event are upgraded to free access when edited during the free launch; the conditional write cannot overwrite a concurrent payment event. Other edits preserve billing rights. There is no client-controlled payment activation.

## Web-only Stripe payments

The web panel charges USD 5.00 once per business through Stripe Checkout, using
the shared Nexode Stripe account and the dedicated `prod_harkai_map_listing`
product. Municipal access is agreed directly and is unrelated to this price.
Existing active free-launch records keep their access. With billing configured,
new records remain pending until a verified payment is received.

Owners first submit their business for content review. Only an approved unpaid
record can create/resume a checkout through
`POST /v1/me/businesses/{id}/checkout`. The server chooses and validates the price,
product, amount, currency, owner and business; clients supply only the business
ID. Persisted checkout attempts share Stripe idempotency keys across replicas.
Only Stripe sessions confirmed expired can be replaced. A missing session after
23 hours fails closed and requires operator reconciliation of Stripe sessions
before the persisted attempt is repaired; do not clear it blindly.

Configuration requires `STRIPE_SECRET_KEY`, `STRIPE_PRICE_ID`,
`STRIPE_WEBHOOK_SECRET`, `STRIPE_SUCCESS_URL` and `STRIPE_CANCEL_URL` together.
Secrets belong only in the deployment environment. Return URLs must use the
Harkai panel origin. A redirect is informational and never activates access.
`POST /v1/business-billing/webhook` verifies the raw Stripe signature and API
version, live/test mode, paid status and USD 5.00 total. Activation must match the
persisted checkout session and owner. Signed unrelated Nexode events are ignored.
Repeated and older events are acknowledged without overwriting newer billing
state. Full refunds revoke access; partial refunds do not. Payment IDs are unique
across businesses. Content approval, payment and the enabled switch must all
permit visibility. Editing preserves payment but resets content review.

The native mobile app contains no purchase SDK, checkout, price or external
purchase link. It lists and manages only existing active one-time or grandfathered
free access, including editing, offers and pausing. Business creation and payment
remain in the web panel. Legacy mobile versions can still create a pending record
through the shared API, but cannot grant free or paid access when billing is on.

## Proximity delivery

Business proximity notices use a registered live-location device and the app’s location/notification permissions. A location update includes accuracy and observed_at. Only accurate fresh GPS within each point radius qualifies; owner notices are suppressed. A persistent per-user/per-point outbox enforces 24-hour cooldown. The response carries business_notices. The app shows a separate promotional notification channel and acknowledges `/v1/me/business-notices/{id}/ack`; a lost response/ack is retried with the same local notification ID. Delivery is scoped to device/session and rechecks public visibility and current proximity. Ordinary safety alert delivery is independent.
