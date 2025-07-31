import React, { useState, useEffect } from 'react';
import { 
  Card, 
  List, 
  Button, 
  Space, 
  Tag, 
  Modal,
  message,
  Tooltip,
  Descriptions
} from 'antd';
import { 
  ToolOutlined,
  PlayCircleOutlined,
  InfoCircleOutlined,
  SettingOutlined
} from '@ant-design/icons';
import ApiService, { Tool } from '../services/api';

const Tools: React.FC = () => {
  const [tools, setTools] = useState<Tool[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedTool, setSelectedTool] = useState<Tool | null>(null);
  const [detailModalVisible, setDetailModalVisible] = useState(false);

  useEffect(() => {
    loadTools();
  }, []);

  const loadTools = async () => {
    setLoading(true);
    try {
      const data = await ApiService.listTools();
      setTools(data);
    } catch (error) {
      message.error('加载工具列表失败');
      console.error('Failed to load tools:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleExecuteTool = async (toolName: string) => {
    try {
      const result = await ApiService.executeTool(toolName, {});
      message.success('工具执行成功');
      console.log('Tool execution result:', result);
    } catch (error) {
      message.error('工具执行失败');
      console.error('Failed to execute tool:', error);
    }
  };

  const handleViewToolDetail = (tool: Tool) => {
    setSelectedTool(tool);
    setDetailModalVisible(true);
  };

  const getToolIcon = (toolName: string) => {
    if (toolName.includes('search')) return '🔍';
    if (toolName.includes('analysis')) return '📊';
    if (toolName.includes('report')) return '📄';
    if (toolName.includes('data')) return '📈';
    return '🔧';
  };

  const getToolColor = (toolName: string) => {
    if (toolName.includes('search')) return 'blue';
    if (toolName.includes('analysis')) return 'green';
    if (toolName.includes('report')) return 'purple';
    if (toolName.includes('data')) return 'orange';
    return 'default';
  };

  return (
    <div>
      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center',
        marginBottom: '16px'
      }}>
        <h2>工具库</h2>
        <Space>
          <Button icon={<SettingOutlined />}>
            工具配置
          </Button>
        </Space>
      </div>

      <Card>
        <List
          loading={loading}
          dataSource={tools}
          grid={{ gutter: 16, xs: 1, sm: 2, md: 3, lg: 4, xl: 4, xxl: 6 }}
          renderItem={(tool) => (
            <List.Item>
              <Card
                hoverable
                style={{ height: '100%' }}
                actions={[
                  <Tooltip title="查看详情">
                    <Button
                      type="text"
                      icon={<InfoCircleOutlined />}
                      onClick={() => handleViewToolDetail(tool)}
                    />
                  </Tooltip>,
                  <Tooltip title="执行工具">
                    <Button
                      type="text"
                      icon={<PlayCircleOutlined />}
                      onClick={() => handleExecuteTool(tool.name)}
                    />
                  </Tooltip>
                ]}
              >
                <Card.Meta
                  avatar={
                    <div style={{ 
                      fontSize: '24px',
                      width: '40px',
                      height: '40px',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      backgroundColor: '#f0f0f0',
                      borderRadius: '8px'
                    }}>
                      {getToolIcon(tool.name)}
                    </div>
                  }
                  title={
                    <Space>
                      <span>{tool.name}</span>
                      <Tag color={getToolColor(tool.name)}>
                        {tool.name.split('_')[0]}
                      </Tag>
                    </Space>
                  }
                  description={
                    <div>
                      <p style={{ marginBottom: '8px' }}>
                        {tool.description}
                      </p>
                      {tool.parameters && Object.keys(tool.parameters).length > 0 && (
                        <div style={{ fontSize: '12px', color: '#666' }}>
                          参数: {Object.keys(tool.parameters).join(', ')}
                        </div>
                      )}
                    </div>
                  }
                />
              </Card>
            </List.Item>
          )}
        />
      </Card>

      <Modal
        title="工具详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="cancel" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>,
          <Button 
            key="execute" 
            type="primary"
            icon={<PlayCircleOutlined />}
            onClick={() => {
              if (selectedTool) {
                handleExecuteTool(selectedTool.name);
                setDetailModalVisible(false);
              }
            }}
          >
            执行工具
          </Button>
        ]}
        width={600}
      >
        {selectedTool && (
          <Descriptions column={1} bordered>
            <Descriptions.Item label="工具名称">
              {selectedTool.name}
            </Descriptions.Item>
            <Descriptions.Item label="描述">
              {selectedTool.description}
            </Descriptions.Item>
            <Descriptions.Item label="参数">
              {selectedTool.parameters ? (
                <pre style={{ fontSize: '12px', backgroundColor: '#f5f5f5', padding: '8px' }}>
                  {JSON.stringify(selectedTool.parameters, null, 2)}
                </pre>
              ) : (
                <span style={{ color: '#999' }}>无参数</span>
              )}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
};

export default Tools; 