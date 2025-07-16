The low-level message CBOR format might be:

- CBOR tag 'grid'
- what i want, consisting of:
    - CID representing the function the counterparty would execute
    - arguments to function
- what i have, consisting of:
    - CID representing the function i would execute
    - arguments to function

If this format is in use, then the kernel would be the counterparty
for all messages; the function to be executed would in turn contain a
function as an argument, which would be the function to be executed
by another agent?

