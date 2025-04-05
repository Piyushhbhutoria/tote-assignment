import axios from 'axios';
import { PluginWithStats } from '../types/plugin';

const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8080/api',
});

export const getPlugins = async (): Promise<PluginWithStats[]> => {
  const response = await api.get('/plugins');
  return response.data;
};

export const updatePluginStatus = async (name: string, isActive: boolean): Promise<void> => {
  await api.patch(`/plugins/${name}/status`, { isActive });
};

export const updatePluginConfig = async (name: string, config: Record<string, any>): Promise<void> => {
  await api.patch(`/plugins/${name}/config`, { config });
}; 
