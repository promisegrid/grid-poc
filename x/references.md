Hashes give us irrefutable references to things that happened earlier
on the time axis (barring collisions).  But traversing forward in time
from a given hash is harder because it's essentially a prediction of
the future from the perspective of the given hash.  Git users assert
"this commit is the next thing on this branch", and IPFS users assert
"this IPNS record is the next version of this thing".  

Git and IPFS are both vulnerable to political fights because they only
support one hash per forward ref -- neither system supports multiple
competing forward refs with the same name.  So if two users disagree
about what the next thing is, they have to resolve that disagreement
out-of-band.  (Hence github's issues and pull requests, and IPFS's 24
hour TTL on IPNS records.)  

Git and IPFS are also vulnerable to hash collisions because they only
support one hash per backward ref -- neither system supports multiple
competing contents for the same hash.  So in the event of a hash
collision, there may be no in-band method of detecting the collision,
let alone resolving it.  

Git and IPFS are vulnerable to misinformation because their references
do not carry in-band risk for the asserter.  Git users can assert a
given commit is the next thing on a given branch, and IPFS users can
assert a given IPNS record is the next version of a given file, but if
their assertions turn out to be false, the only consequence is

XXX


tag or branch points to a given commit, and IPFS users can assert an
IPNS record points to a given hash, but if the reader disagrees with



refs are assertions of history without intrinsic proof.  If the reader
disagrees with the assertion, they have no choice but to follow the


PromiseGrid reduces these vulnerabilities by allowing multiple forward
or backward refs per name or hash, and by wrapping references in
promises so the reader has more information and can choose among
multiple competing refs when they exist.

For example, a backward ref in PG is a promise that "this hash
represents this content", and a forward ref is a promise that "hash B
comes after hash A".  If the reader disagrees with the promise or
believes it to be broken, they can choose a different ref to follow
instead.  This is in contrast to Git and IPFS, where the reader has no
choice but to follow the one ref that exists for a given name or hash.

From the perspective of the issuer, PG promises carry in-band risk


By comparison, PG agents make promises that "this thing happens next",
but these promises carry the risk that, if consensus moves away from
the promise, the promise will be considered broken by readers,
resulting in lower point value for the issuer.  So there's incentive
for issuers to do their homework before making an assertion.





include in-band risk -- promise breakage means lower point value over
time.

Restating all that:  Hashes give us irrefutable references to things
that happened earlier on the time axis (barring collisions).  But
forward traversals in time from a given hash are refutable because
they are essentially predictions of the future from the perspective of
the given hash.  Git users assert branch commits, IPFS users assert
IPNS records, and neither of these carry in-band risk for the asserter
-- the risk of asserting a bad ref is all out-of-band, political, or
somewhere else in the system. 

By comparison, PG refs are promises that include in-band risk --
promise breakage leads to lower agent point value, so there's
incentive for issuers to do their homework before making an assertion.  


But Git and IPFS are
both vulnerable to political fights and hash collisions because they
only support one hash per ref name, while PG reduces that
vulnerability by wrapping references in promises so the reader has
more information and can choose among multiple competing refs with the
same name.
