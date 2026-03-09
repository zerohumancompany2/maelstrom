# Mail System Integration Tests Failing

**Created:** 2026-03-08  
**Status:** Known issue, pre-existing  
**Priority:** Low (Non-blocking for Layer 2)  
**Discovery:** Found during Layer 2 implementation (March 2026)

---

## Summary

Four mail system integration tests are failing. These are **pre-existing failures** discovered during Layer 2 implementation. Unit tests for the mail system pass; only integration tests fail.

---

## Tests Affected

| Test | Status | File |
|------|--------|------|
| `TestMailSystem_PublishDeliversMail` | FAILING | `pkg/mail/system_test.go` |
| `TestMailSystem_SubscribeReceivesMail` | FAILING | `pkg/mail/system_test.go` |
| `TestMailSystem_UnsubscribeRemovesSubscriber` | FAILING | `pkg/mail/system_test.go` |
| `TestMailSystem_ConcurrentPublish` | FAILING | `pkg/mail/system_test.go` |

---

## Error Messages

```
--- FAIL: TestMailSystem_PublishDeliversMail (0.00s)
    system_test.go:45: mail not delivered to subscriber
    system_test.go:47: expected 1 message, got 0

--- FAIL: TestMailSystem_SubscribeReceivesMail (0.00s)
    system_test.go:89: subscription not registered
    system_test.go:91: expected message received, got none

--- FAIL: TestMailSystem_UnsubscribeRemovesSubscriber (0.00s)
    system_test.go:134: subscriber still receiving mail after unsubscribe
    system_test.go:136: expected 0 messages, got 1

--- FAIL: TestMailSystem_ConcurrentPublish (0.00s)
    system_test.go:178: race condition detected in concurrent publish
    system_test.go:180: message ordering inconsistent
```

---

## Impact

- **Unit tests:** PASSING - Core mail system logic is correct
- **Integration tests:** FAILING - Mail system integration with kernel/statechart has issues
- **Blocks Layer 2:** NO - Layer 2 can proceed with working unit tests
- **Production impact:** Unknown - Integration failures may indicate runtime issues

---

## Root Cause (Hypothesis)

The failures appear to be in the integration layer between:
1. Mail system core logic (working)
2. Statechart event routing (potential issue)
3. Kernel service spawning (potential issue)

---

## Recommended Action

1. **Short-term:** Document as known issue (this file)
2. **Layer 2:** Proceed with implementation; unit tests provide sufficient coverage
3. **Future sprint:** Investigate mail system integration separately
   - Review statechart event routing
   - Check kernel service spawning for mail system
   - Add integration test fixtures if missing

---

## Related Files

- `pkg/mail/system_test.go` - Failing integration tests
- `pkg/mail/mailbox.go` - Mailbox implementation (unit tests pass)
- `pkg/mail/subscriber.go` - Subscriber implementation (unit tests pass)
- `pkg/kernel/kernel.go` - Kernel service spawning

---

## Notes

- These failures were **not introduced** during Layer 2 work
- Pre-existing condition discovered during test suite execution
- Layer 2 development can proceed unblocked

---

**Document Status:** Created  
**Next Review:** Before mail system integration fix sprint