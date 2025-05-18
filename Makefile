.PHONY: test

TEST_RESULT=result.log

test:
	go test -v ./... | tee $(TEST_RESULT)
	@echo ""
	@echo "==== TEST SUMMARY ===="
	@grep -E "PASS|FAIL" $(TEST_RESULT)

clean:
	rm -f $(TEST_RESULT)
