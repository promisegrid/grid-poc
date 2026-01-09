# TODO 002: Decide what to do with `imposition` (request) mtype

## Context

The example Promise Theory protocol (`pt_pCID`) currently sketches:

- `0` promise (offer about self)
- `1` imposition (request)
- `2` assessment (receipt/observation)

We need to decide whether `imposition` belongs in the minimal long-horizon
protocol, or whether requests should instead be represented as conditional
promises (e.g., bids/asks/offers) plus independent assessments.

## Questions to Answer

- Do we want a first-class “request” primitive on the wire, or keep the wire
  vocabulary to promises + assessments only?
- If we keep `imposition`, what are the normative semantics?
  - It MUST NOT imply obligation; it is always ignorable.
  - What limits/filters should nodes apply to unsolicited impositions?
- If we drop `imposition`, what is the standard pattern for:
  - RPC-like calls (query/fetch/subscribe)?
  - Coordination requests (merge/consensus formation)?
  - Market “asks” vs “bids” (conditional promises as price language)?

## Deliverables

- Record the decision (keep vs drop) with rationale and primary use cases.
- Update the example protocol in:
  - `x/rfc/draft-promisegrid.md` (Section 6)
  - `x/wire/wire.md` (Section 7.1)
