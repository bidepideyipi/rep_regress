package com.rtp.flink.deserializer;

import com.alibaba.fastjson2.JSON;
import com.rtp.flink.model.GameLogDetail;
import org.apache.flink.api.common.serialization.DeserializationSchema;
import org.apache.flink.api.common.typeinfo.TypeInformation;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

/**
 * RocketMQ消息反序列化器
 * 将RocketMQ消息体解析为GameLogDetail对象
 * 参考 rtp-processor 中的 JSON 解析逻辑
 */
public class GameLogDeserializer implements DeserializationSchema<GameLogDetail> {

    private static final Logger LOG = LoggerFactory.getLogger(GameLogDeserializer.class);

    @Override
    public GameLogDetail deserialize(byte[] message) throws IOException {
        if (message == null || message.length == 0) {
            LOG.warn("收到空消息，跳过处理");
            return null;
        }

        try {
            String jsonStr = new String(message, StandardCharsets.UTF_8);
            return JSON.parseObject(jsonStr, GameLogDetail.class);
        } catch (Exception e) {
            LOG.error("JSON解析失败: {}", new String(message, StandardCharsets.UTF_8), e);
            // 返回null表示该消息无法解析，将被过滤
            return null;
        }
    }

    @Override
    public boolean isEndOfStream(GameLogDetail nextElement) {
        return false;
    }

    @Override
    public TypeInformation<GameLogDetail> getProducedType() {
        return TypeInformation.of(GameLogDetail.class);
    }
}
