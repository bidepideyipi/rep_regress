package com.rtp.flink.config;

import com.alibaba.fastjson2.JSON;
import com.alibaba.nacos.api.NacosFactory;
import com.alibaba.nacos.api.config.ConfigService;
import com.alibaba.nacos.api.exception.NacosException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Nacos配置加载器
 * 从Nacos配置中心加载应用配置
 *
 * 参考 rtp-processor 中的 Nacos 配置实现
 */
public class NacosConfigLoader {

    private static final Logger LOG = LoggerFactory.getLogger(NacosConfigLoader.class);
    private static final String DEFAULT_DATA_ID = "app-config";
    private static final String DEFAULT_GROUP = "DEFAULT_GROUP";
    private static final int TIMEOUT_MS = 5000;

    private final ConfigService configService;

    /**
     * 创建Nacos配置加载器
     *
     * @param serverAddr Nacos服务器地址 (例如: "127.0.0.1:8848")
     * @param namespace  命名空间ID (空字符串表示public命名空间)
     */
    public NacosConfigLoader(String serverAddr, String namespace) throws NacosException {
        LOG.info("连接Nacos: {}", serverAddr);

        java.util.Properties properties = new java.util.Properties();
        properties.put("serverAddr", serverAddr);
        if (namespace != null && !namespace.isEmpty()) {
            properties.put("namespace", namespace);
        }

        this.configService = NacosFactory.createConfigService(properties);

        // 测试连接
        configService.getServerStatus();
        LOG.info("Nacos连接成功");
    }

    /**
     * 加载配置
     *
     * @param dataId 配置文件的Data ID (默认: "app-config")
     * @param group  配置分组 (默认: "DEFAULT_GROUP")
     * @return 应用配置对象
     * @throws NacosException 加载失败
     */
    public AppConfig loadConfig(String dataId, String group) throws NacosException {
        String actualDataId = dataId != null ? dataId : DEFAULT_DATA_ID;
        String actualGroup = group != null ? group : DEFAULT_GROUP;

        LOG.info("从Nacos加载配置: {} @ {}", actualDataId, actualGroup);

        String configContent = configService.getConfig(actualDataId, actualGroup, TIMEOUT_MS);

        if (configContent == null || configContent.trim().isEmpty()) {
            throw new IllegalStateException("Nacos配置为空: " + actualDataId);
        }

        LOG.info("配置加载成功:\n{}", configContent);
        return JSON.parseObject(configContent, AppConfig.class);
    }

    /**
     * 关闭配置服务
     */
    public void close() {
        // Nacos Client SDK 没有提供明确的关闭方法
    }
}
