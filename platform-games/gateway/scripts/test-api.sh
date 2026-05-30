#!/bin/bash

GATEWAY_URL="http://localhost:8080"

echo "=== Gateway API 测试 ==="
echo ""

echo "1. 健康检查"
curl -s "$GATEWAY_URL/health" | jq '.'
echo ""

echo "2. 游戏访问授权"
curl -s -X POST "$GATEWAY_URL/v1/auth/game-access" \
  -H "Content-Type: application/json" \
  -d '{
    "merchant_id": "merchant_001",
    "user_id": "user_12345",
    "game_id": "slot_game_v1",
    "language": "zh-CN",
    "currency": "CNY",
    "timestamp": '$(date +%s%N)',
    "nonce": "a1b2c3d4e5f6g7h8",
    "signature": "test_signature"
  }' | jq '.'
echo ""

echo "3. 令牌验证"
curl -s -X POST "$GATEWAY_URL/v1/auth/verify-token" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "test_token",
    "timestamp": '$(date +%s%N)',
    "signature": "test_signature"
  }' | jq '.'
echo ""

echo "4. 用户信息"
curl -s "$GATEWAY_URL/v1/user/info" \
  -H "Authorization: Bearer test_token" | jq '.'
echo ""

echo "5. 余额查询"
curl -s "$GATEWAY_URL/v1/finance/balance" \
  -H "Authorization: Bearer test_token" | jq '.'
echo ""

echo "=== 测试完成 ==="
