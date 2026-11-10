# Business points and permanent publications

Incident markers last three hours, including older pet/place records. MongoDB retains the incident, its moderated photo and comments; `/v1/publications` lists non-hidden records for the public website without authentication or geographic restrictions. Hiding a report removes its public detail, list and photo access.

Business points are separate records. Owners create/edit through `/v1/me/businesses`, admins approve/reject through `/v1/admin/businesses/{id}/moderation`. Public queries, photo access and proximity notices require enabled + approved + active access. During the free launch, the server grants payment_mode=free when no billing adapter is configured. No card or checkout is required. One-time paid records still require a valid paid_at; legacy subscriptions require future paid_until. Edits reset review. Unpaid legacy drafts with no payment event are upgraded to free access when edited during the free launch; the conditional write cannot overwrite a concurrent payment event. Other edits preserve billing rights. There is no client-controlled payment activation.

## Stripe remains pending

The production environment has credentials for the shared Nexode Stripe account
and a dedicated Harkai product with one USD 5.00 one-time price. Configuration uses
`STRIPE_SECRET_KEY`, `STRIPE_PRODUCT_ID`, `STRIPE_PRICE_ID`, `STRIPE_CURRENCY`,
`STRIPE_PRICE_CENTS`, `STRIPE_SUCCESS_URL` and `STRIPE_CANCEL_URL`. Secrets stay in
the deployment environment. This prepared configuration does not itself enable
collection; the checkout adapter and signed webhook flow below are still required.

`Service.Billing` is intentionally nil in production. `/v1/business-billing/webhook` returns 503 and meta payments=false until a real adapter is wired. To enable payments, implement `BusinessBillingVerifier`: verify Stripe's raw signed body, retrieve authoritative subscription/invoice status, validate the configured product/price/currency and server-side point/owner binding, then emit BusinessEntitlement. Never trust checkout return URLs or arbitrary client metadata as proof of payment. Map paid periods and cancellation policy explicitly. Configure prices, credentials and webhook secret outside source control. The store applies events idempotently and rejects older events and mismatched owners. No checkout UI pretends that payment succeeded.

## Proximity delivery

Business proximity notices use a registered live-location device and the app’s location/notification permissions. A location update includes accuracy and observed_at. Only accurate fresh GPS within each point radius qualifies; owner notices are suppressed. A persistent per-user/per-point outbox enforces 24-hour cooldown. The response carries business_notices. The app shows a separate promotional notification channel and acknowledges `/v1/me/business-notices/{id}/ack`; a lost response/ack is retried with the same local notification ID. Delivery is scoped to device/session and rechecks public visibility and current proximity. Ordinary safety alert delivery is independent.
