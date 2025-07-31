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
  Tooltip,
  Badge
} from 'antd';
import { 
  PlusOutlined, 
  MessageOutlined, 
  DeleteOutlined,
  EyeOutlined,
  EditOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import ApiService, { Session, Agent, CreateSessionRequest } from '../services/api';

const { Option } = Select;
const { TextArea } = Input;

const Sessions: React.FC = () => {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();

  useEffect(() => {
    loadSessions();
    loadAgents();
  }, []);

  const loadSessions = async () => {
    setLoading(true);
    try {
      const data = await ApiService.listSessions();
      setSessions(data);
    } catch (error) {
      message.error('加载会话列表失败');
      console.error('Failed to load sessions:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadAgents = async () => {
    try {
      const data = await ApiService.listAgents();
      setAgents(data);
    } catch (error) {
      console.error('Failed to load agents:', error);
    }
  };

  const handleCreateSession = async (values: CreateSessionRequest) => {
    try {
      await ApiService.createSession(values);
      message.success('会话创建成功');
      setCreateModalVisible(false);
      form.resetFields();
      loadSessions();
    } catch (error) {
      message.error('创建会话失败');
      console.error('Failed to create session:', error);
    }
  };

  const handleDeleteSession = async (sessionId: string) => {
    try {
      await ApiService.deleteSession(sessionId);
      message.success('会话删除成功');
      loadSessions();
    } catch (error) {
      message.error('删除会话失败');
      console.error('Failed to delete session:', error);
    }
  };

  const handleViewSession = (sessionId: string) => {
    navigate(`/sessions/${sessionId}`);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'green';
      case 'completed':
        return 'blue';
      case 'paused':
        return 'orange';
      case 'cancelled':
        return 'red';
      default:
        return 'default';
    }
  };

  const columns = [
    {
      title: '会话名称',
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
      title: '关联智能体',
      dataIndex: 'agent_id',
      key: 'agent_id',
      render: (agentId: string) => {
        const agent = agents.find(a => a.id === agentId);
        return agent ? (
          <Tag color="blue">{agent.name}</Tag>
        ) : (
          <Tag color="default">未知</Tag>
        );
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Badge 
          status={status === 'active' ? 'processing' : 'default'} 
          text={
            <Tag color={getStatusColor(status)}>
              {status}
            </Tag>
          }
        />
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
      render: (_: any, record: Session) => (
        <Space>
          <Tooltip title="查看会话">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleViewSession(record.id)}
            />
          </Tooltip>
          
          <Tooltip title="开始对话">
            <Button
              type="text"
              icon={<MessageOutlined />}
              onClick={() => navigate(`/sessions/${record.id}/chat`)}
            />
          </Tooltip>
          
          <Popconfirm
            title="确定要删除这个会话吗？"
            onConfirm={() => handleDeleteSession(record.id)}
          >
            <Tooltip title="删除会话">
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
              />
            </Tooltip>
          </Popconfirm>
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
        <h2>研究会话</h2>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setCreateModalVisible(true)}
        >
          创建会话
        </Button>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={sessions}
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
        title="创建研究会话"
        open={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        footer={null}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateSession}
        >
          <Form.Item
            name="name"
            label="会话名称"
            rules={[{ required: true, message: '请输入会话名称' }]}
          >
            <Input placeholder="请输入会话名称" />
          </Form.Item>

          <Form.Item
            name="description"
            label="描述"
            rules={[{ required: true, message: '请输入描述' }]}
          >
            <TextArea 
              rows={3} 
              placeholder="请输入会话描述"
            />
          </Form.Item>

          <Form.Item
            name="agent_id"
            label="选择智能体"
            rules={[{ required: true, message: '请选择智能体' }]}
          >
            <Select placeholder="请选择智能体">
              {agents.map(agent => (
                <Option key={agent.id} value={agent.id}>
                  {agent.name} ({agent.type})
                </Option>
              ))}
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

export default Sessions; 