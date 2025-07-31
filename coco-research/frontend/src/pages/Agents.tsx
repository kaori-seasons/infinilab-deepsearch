import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Table, 
  Button, 
  Space, 
  Tag, 
  Modal, 
  Form, 
  Input, 
  Select,
  message,
  Popconfirm,
  Tooltip
} from 'antd';
import { 
  PlusOutlined, 
  PlayCircleOutlined, 
  StopOutlined, 
  EyeOutlined,
  DeleteOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import ApiService, { Agent, CreateAgentRequest } from '../services/api';

const { Option } = Select;
const { TextArea } = Input;

const Agents: React.FC = () => {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();

  useEffect(() => {
    loadAgents();
  }, []);

  const loadAgents = async () => {
    setLoading(true);
    try {
      const data = await ApiService.listAgents();
      setAgents(data);
    } catch (error) {
      message.error('加载智能体列表失败');
      console.error('Failed to load agents:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateAgent = async (values: CreateAgentRequest) => {
    try {
      await ApiService.createAgent(values);
      message.success('智能体创建成功');
      setCreateModalVisible(false);
      form.resetFields();
      loadAgents();
    } catch (error) {
      message.error('创建智能体失败');
      console.error('Failed to create agent:', error);
    }
  };

  const handleStopAgent = async (agentId: string) => {
    try {
      await ApiService.stopAgent(agentId);
      message.success('智能体已停止');
      loadAgents();
    } catch (error) {
      message.error('停止智能体失败');
      console.error('Failed to stop agent:', error);
    }
  };

  const handleViewAgent = (agentId: string) => {
    navigate(`/agents/${agentId}`);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
        return 'green';
      case 'stopped':
        return 'red';
      case 'idle':
        return 'blue';
      default:
        return 'default';
    }
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => <strong>{text}</strong>,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag color="blue">{type}</Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'state',
      key: 'state',
      render: (state: string) => (
        <Tag color={getStatusColor(state)}>
          {state}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: Agent) => (
        <Space>
          <Tooltip title="查看详情">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleViewAgent(record.id)}
            />
          </Tooltip>
          
          {record.state === 'running' && (
            <Popconfirm
              title="确定要停止这个智能体吗？"
              onConfirm={() => handleStopAgent(record.id)}
            >
              <Tooltip title="停止智能体">
                <Button
                  type="text"
                  danger
                  icon={<StopOutlined />}
                />
              </Tooltip>
            </Popconfirm>
          )}
          
          {record.state !== 'running' && (
            <Tooltip title="启动智能体">
              <Button
                type="text"
                icon={<PlayCircleOutlined />}
                onClick={() => navigate(`/agents/${record.id}/execute`)}
              />
            </Tooltip>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center',
        marginBottom: '16px'
      }}>
        <h2>智能体管理</h2>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setCreateModalVisible(true)}
        >
          创建智能体
        </Button>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={agents}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>

      <Modal
        title="创建智能体"
        open={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        footer={null}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateAgent}
        >
          <Form.Item
            name="name"
            label="智能体名称"
            rules={[{ required: true, message: '请输入智能体名称' }]}
          >
            <Input placeholder="请输入智能体名称" />
          </Form.Item>

          <Form.Item
            name="description"
            label="描述"
            rules={[{ required: true, message: '请输入描述' }]}
          >
            <TextArea 
              rows={3} 
              placeholder="请输入智能体描述"
            />
          </Form.Item>

          <Form.Item
            name="type"
            label="类型"
            rules={[{ required: true, message: '请选择类型' }]}
          >
            <Select placeholder="请选择智能体类型">
              <Option value="research">研究智能体</Option>
              <Option value="analysis">分析智能体</Option>
              <Option value="report">报告智能体</Option>
            </Select>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                创建
              </Button>
              <Button onClick={() => setCreateModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Agents;