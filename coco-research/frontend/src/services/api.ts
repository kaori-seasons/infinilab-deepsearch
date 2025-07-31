import axios from 'axios';
import type { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from 'axios';

// API基础配置
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

// 创建axios实例
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // 可以在这里添加认证token
    const token = localStorage.getItem('auth_token');
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error: any) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    return response;
  },
  (error: any) => {
    // 统一错误处理
    if (error.response?.status === 401) {
      // 未授权，跳转到登录页
      localStorage.removeItem('auth_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// 类型定义
export interface Agent {
  id: string;
  name: string;
  description: string;
  type: string;
  state: string;
  created_at: string;
  updated_at: string;
}

export interface Task {
  id: string;
  agent_id: string;
  query: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  result?: string;
  created_at: string;
  updated_at: string;
}

export interface Session {
  id: string;
  name: string;
  description: string;
  agent_id: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface Tool {
  name: string;
  description: string;
  parameters: Record<string, any>;
}

export interface CreateAgentRequest {
  name: string;
  description: string;
  type: string;
}

export interface ExecuteTaskRequest {
  query: string;
  parameters?: Record<string, any>;
}

export interface CreateSessionRequest {
  name: string;
  description: string;
  agent_id: string;
}

// API服务类
export class ApiService {
  // 智能体相关API
  static async listAgents(): Promise<Agent[]> {
    const response = await apiClient.get('/agents');
    return response.data.data;
  }

  static async getAgent(id: string): Promise<Agent> {
    const response = await apiClient.get(`/agents/${id}`);
    return response.data.data;
  }

  static async createAgent(data: CreateAgentRequest): Promise<Agent> {
    const response = await apiClient.post('/agents', data);
    return response.data.data;
  }

  static async executeTask(agentId: string, data: ExecuteTaskRequest): Promise<Task> {
    const response = await apiClient.post(`/agents/${agentId}/execute`, data);
    return response.data.data;
  }

  static async getTaskStatus(taskId: string): Promise<Task> {
    const response = await apiClient.get(`/agents/tasks/${taskId}`);
    return response.data.data;
  }

  static async stopAgent(agentId: string): Promise<void> {
    await apiClient.post(`/agents/${agentId}/stop`);
  }

  // 会话相关API
  static async listSessions(): Promise<Session[]> {
    const response = await apiClient.get('/sessions');
    return response.data.data;
  }

  static async getSession(id: string): Promise<Session> {
    const response = await apiClient.get(`/sessions/${id}`);
    return response.data.data;
  }

  static async createSession(data: CreateSessionRequest): Promise<Session> {
    const response = await apiClient.post('/sessions', data);
    return response.data.data;
  }

  static async updateSession(id: string, data: Partial<Session>): Promise<Session> {
    const response = await apiClient.put(`/sessions/${id}`, data);
    return response.data.data;
  }

  static async deleteSession(id: string): Promise<void> {
    await apiClient.delete(`/sessions/${id}`);
  }

  // 任务相关API
  static async listTasks(): Promise<Task[]> {
    const response = await apiClient.get('/tasks');
    return response.data.data;
  }

  static async getTask(id: string): Promise<Task> {
    const response = await apiClient.get(`/tasks/${id}`);
    return response.data.data;
  }

  static async updateTask(id: string, data: Partial<Task>): Promise<Task> {
    const response = await apiClient.put(`/tasks/${id}`, data);
    return response.data.data;
  }

  static async deleteTask(id: string): Promise<void> {
    await apiClient.delete(`/tasks/${id}`);
  }

  // 工具相关API
  static async listTools(): Promise<Tool[]> {
    const response = await apiClient.get('/tools');
    return response.data.data;
  }

  static async getTool(name: string): Promise<Tool> {
    const response = await apiClient.get(`/tools/${name}`);
    return response.data.data;
  }

  static async executeTool(name: string, input: Record<string, any>): Promise<any> {
    const response = await apiClient.post(`/tools/${name}/execute`, { input });
    return response.data.data;
  }

  // 健康检查
  static async healthCheck(): Promise<any> {
    const response = await apiClient.get('/health');
    return response.data;
  }
}

export default ApiService;