#!/usr/bin/env bash
# Test runner for business metric reordering scenarios
# Runs each config through apply+plan to detect drift

set -euo pipefail

TERRAFORM=/tmp/terraform
MOCKAPI_PID=""
MOCKAPI_PORT=9090
TEST_DIR=/workspace/mockapi/testconfigs
WORK_DIR=/tmp/bm_tests
PASS=0
FAIL=0
FAILURES=()

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

cleanup() {
    if [ -n "$MOCKAPI_PID" ]; then
        kill "$MOCKAPI_PID" 2>/dev/null || true
        wait "$MOCKAPI_PID" 2>/dev/null || true
    fi
}
trap cleanup EXIT

start_mockapi() {
    local mode="${1:-reverse}"
    cleanup  # Stop any existing instance
    
    local flags="-reverse=false -shuffle=false"
    case "$mode" in
        reverse) flags="-reverse=true" ;;
        shuffle) flags="-shuffle=true -reverse=false" ;;
        passthrough) flags="-reverse=false -shuffle=false" ;;
    esac
    
    echo "Starting mock API (mode: $mode)..."
    /workspace/mockapi/mockapi $flags -port=$MOCKAPI_PORT > /tmp/mockapi.log 2>&1 &
    MOCKAPI_PID=$!
    sleep 0.5
    
    # Verify it's running
    if ! kill -0 "$MOCKAPI_PID" 2>/dev/null; then
        echo "Mock API failed to start!"
        cat /tmp/mockapi.log
        exit 1
    fi
    echo "Mock API running (PID: $MOCKAPI_PID)"
}

run_test() {
    local test_name="$1"
    local config_file="$2"
    local extra_steps="${3:-}"  # Optional extra step function name
    local mode="${4:-reverse}"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "TEST: $test_name [mode=$mode]"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    local test_dir="$WORK_DIR/${test_name//[^a-zA-Z0-9]/_}"
    rm -rf "$test_dir"
    mkdir -p "$test_dir"
    cp "$config_file" "$test_dir/main.tf"
    
    # Start mock API for this test
    start_mockapi "$mode"
    
    local test_passed=true
    
    # Step 1: Init
    echo "→ terraform init..."
    if ! $TERRAFORM -chdir="$test_dir" init -upgrade -no-color > "$test_dir/init.log" 2>&1; then
        echo -e "${RED}  FAIL: init failed${NC}"
        cat "$test_dir/init.log" | tail -20
        test_passed=false
    fi
    
    if $test_passed; then
        # Step 2: Apply
        echo "→ terraform apply..."
        if ! $TERRAFORM -chdir="$test_dir" apply -auto-approve -no-color > "$test_dir/apply.log" 2>&1; then
            echo -e "${RED}  FAIL: apply failed${NC}"
            cat "$test_dir/apply.log" | tail -30
            test_passed=false
        else
            echo -e "${GREEN}  apply: OK${NC}"
            # Show what was created
            grep -E "vantage_business_metric\..* created" "$test_dir/apply.log" | head -5
        fi
    fi
    
    if $test_passed; then
        # Step 3: Plan (should show no changes)
        echo "→ terraform plan (checking for drift after apply)..."
        local plan_output
        plan_output=$($TERRAFORM -chdir="$test_dir" plan -no-color 2>&1 || true)
        echo "$plan_output" > "$test_dir/plan_after_apply.log"
        
        if echo "$plan_output" | grep -q "No changes\|no changes"; then
            echo -e "${GREEN}  plan after apply: No drift ✓${NC}"
        else
            echo -e "${RED}  FAIL: Plan shows changes after apply (REORDERING DETECTED!)${NC}"
            echo "$plan_output" | grep -A5 -B2 "must be replaced\|will be updated\|~ " | head -50
            test_passed=false
        fi
    fi
    
    # Run extra steps if defined
    if $test_passed && [ -n "$extra_steps" ]; then
        "$extra_steps" "$test_dir"
        test_passed=$?
        [ $test_passed -eq 0 ] && test_passed=true || test_passed=false
    fi
    
    if $test_passed; then
        echo -e "${GREEN}  PASS: $test_name${NC}"
        PASS=$((PASS+1))
    else
        echo -e "${RED}  FAIL: $test_name${NC}"
        FAIL=$((FAIL+1))
        FAILURES+=("$test_name [mode=$mode]")
        echo "Mock API log:"
        tail -20 /tmp/mockapi.log
    fi
}

run_test_with_update() {
    local test_name="$1"
    local config_v1="$2"
    local config_v2="$3"
    local mode="${4:-reverse}"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "TEST (update): $test_name [mode=$mode]"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    local test_dir="$WORK_DIR/${test_name//[^a-zA-Z0-9]/_}"
    rm -rf "$test_dir"
    mkdir -p "$test_dir"
    
    start_mockapi "$mode"
    
    local test_passed=true
    
    # Init
    cp "$config_v1" "$test_dir/main.tf"
    echo "→ terraform init..."
    if ! $TERRAFORM -chdir="$test_dir" init -upgrade -no-color > "$test_dir/init.log" 2>&1; then
        echo -e "${RED}  FAIL: init failed${NC}"
        test_passed=false
    fi
    
    if $test_passed; then
        # Apply v1
        echo "→ terraform apply v1..."
        if ! $TERRAFORM -chdir="$test_dir" apply -auto-approve -no-color > "$test_dir/apply_v1.log" 2>&1; then
            echo -e "${RED}  FAIL: apply v1 failed${NC}"
            cat "$test_dir/apply_v1.log" | tail -20
            test_passed=false
        else
            echo -e "${GREEN}  apply v1: OK${NC}"
        fi
    fi
    
    if $test_passed; then
        # Update config
        cp "$config_v2" "$test_dir/main.tf"
        
        # Plan update (check for unexpected drift)
        echo "→ terraform plan v2..."
        local plan_output
        plan_output=$($TERRAFORM -chdir="$test_dir" plan -no-color 2>&1 || true)
        echo "$plan_output" > "$test_dir/plan_v2.log"
        
        # Check that the plan only shows the expected change (title update)
        local unexpected_drift=false
        if echo "$plan_output" | grep -q "~ cost_report_tokens_with_metadata"; then
            echo -e "${RED}  FAIL: Plan shows unexpected cost_report_tokens reordering!${NC}"
            echo "$plan_output" | grep -A10 "cost_report_tokens_with_metadata"
            unexpected_drift=true
        fi
        if echo "$plan_output" | grep -qE "^[[:space:]]+~ values[[:space:]]*=|^[[:space:]]+~ forecasted_values[[:space:]]*="; then
            echo -e "${RED}  FAIL: Plan shows unexpected values/forecasted_values reordering!${NC}"
            echo "$plan_output" | grep -A5 "values"
            unexpected_drift=true
        fi
        
        if $unexpected_drift; then
            test_passed=false
        fi
        
        # Apply update
        echo "→ terraform apply v2..."
        if ! $TERRAFORM -chdir="$test_dir" apply -auto-approve -no-color > "$test_dir/apply_v2.log" 2>&1; then
            echo -e "${RED}  FAIL: apply v2 failed${NC}"
            cat "$test_dir/apply_v2.log" | tail -20
            test_passed=false
        else
            echo -e "${GREEN}  apply v2: OK${NC}"
        fi
    fi
    
    if $test_passed; then
        # Final plan check - should be no drift
        echo "→ terraform plan after update (checking for drift)..."
        local final_plan
        final_plan=$($TERRAFORM -chdir="$test_dir" plan -no-color 2>&1 || true)
        echo "$final_plan" > "$test_dir/plan_final.log"
        
        if echo "$final_plan" | grep -q "No changes\|no changes"; then
            echo -e "${GREEN}  final plan: No drift ✓${NC}"
        else
            echo -e "${RED}  FAIL: Final plan shows changes (persistent reordering!)${NC}"
            echo "$final_plan" | grep -E "~|->|must be replaced" | head -30
            test_passed=false
        fi
    fi
    
    if $test_passed; then
        echo -e "${GREEN}  PASS: $test_name${NC}"
        PASS=$((PASS+1))
    else
        echo -e "${RED}  FAIL: $test_name${NC}"
        FAIL=$((FAIL+1))
        FAILURES+=("$test_name [mode=$mode] [update-test]")
        echo "Mock API log:"
        tail -30 /tmp/mockapi.log
    fi
}

# Generate a title-update config variant
make_title_update_config() {
    local original_config="$1"
    local new_title="$2"
    local output_file="$3"
    sed "s/title = \".*\"/title = \"$new_title\"/" "$original_config" > "$output_file"
}

mkdir -p "$WORK_DIR"

echo "====================================="
echo "Business Metric Reordering Test Suite"
echo "====================================="
echo ""

# ─────────────────────────────────────────────
# Group 1: Basic apply+plan with API returning reversed order
# ─────────────────────────────────────────────
echo -e "\n${YELLOW}=== GROUP 1: Apply+Plan (API returns reversed order) ===${NC}"

run_test "3_tokens_reversed" "$TEST_DIR/01_basic_cost_report_tokens.tf" "" "reverse"
run_test "5_tokens_no_labels_reversed" "$TEST_DIR/02_five_tokens_no_labels.tf" "" "reverse"
run_test "values_and_tokens_reversed" "$TEST_DIR/03_values_and_tokens.tf" "" "reverse"
run_test "labeled_values_reversed" "$TEST_DIR/04_values_with_labels.tf" "" "reverse"
run_test "forecasted_values_reversed" "$TEST_DIR/05_forecasted_values.tf" "" "reverse"
run_test "8_tokens_reversed" "$TEST_DIR/06_many_tokens.tf" "" "reverse"
run_test "omitted_label_filter_reversed" "$TEST_DIR/07_no_label_filter_key.tf" "" "reverse"
run_test "mixed_filters_reversed" "$TEST_DIR/08_mixed_label_filters.tf" "" "reverse"

# ─────────────────────────────────────────────
# Group 2: Apply+Plan with shuffled (rotated) order
# ─────────────────────────────────────────────
echo -e "\n${YELLOW}=== GROUP 2: Apply+Plan (API rotates token order) ===${NC}"

run_test "3_tokens_shuffled" "$TEST_DIR/01_basic_cost_report_tokens.tf" "" "shuffle"
run_test "5_tokens_shuffled" "$TEST_DIR/02_five_tokens_no_labels.tf" "" "shuffle"
run_test "8_tokens_shuffled" "$TEST_DIR/06_many_tokens.tf" "" "shuffle"

# ─────────────────────────────────────────────
# Group 3: Update scenarios (title change only, tokens should not drift)
# ─────────────────────────────────────────────
echo -e "\n${YELLOW}=== GROUP 3: Title-only update (tokens must not reorder) ===${NC}"

# Create v1 configs (with different titles)
for config in "$TEST_DIR"/0{1,2,3,4,5,6,7,8}_*.tf; do
    name=$(basename "$config" .tf)
    v1="$WORK_DIR/${name}_v1.tf"
    v2="$WORK_DIR/${name}_v2.tf"
    make_title_update_config "$config" "Test v1 - $(basename $config .tf)" "$v1"
    make_title_update_config "$config" "Test v2 - $(basename $config .tf)" "$v2"
    run_test_with_update "update_${name}" "$v1" "$v2" "reverse"
done

# ─────────────────────────────────────────────
# Group 4: Passthrough (API returns same order) - should always pass
# ─────────────────────────────────────────────
echo -e "\n${YELLOW}=== GROUP 4: Passthrough (baseline - API preserves order) ===${NC}"

run_test "3_tokens_passthrough" "$TEST_DIR/01_basic_cost_report_tokens.tf" "" "passthrough"
run_test "8_tokens_passthrough" "$TEST_DIR/06_many_tokens.tf" "" "passthrough"

# ─────────────────────────────────────────────
# Summary
# ─────────────────────────────────────────────
echo ""
echo "====================================="
echo "TEST SUMMARY"
echo "====================================="
echo -e "${GREEN}PASSED: $PASS${NC}"
echo -e "${RED}FAILED: $FAIL${NC}"

if [ ${#FAILURES[@]} -gt 0 ]; then
    echo ""
    echo "Failed tests:"
    for f in "${FAILURES[@]}"; do
        echo -e "  ${RED}✗ $f${NC}"
    done
    exit 1
else
    echo -e "\n${GREEN}All tests passed!${NC}"
fi
