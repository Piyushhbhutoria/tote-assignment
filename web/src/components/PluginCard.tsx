import { Switch } from '@headlessui/react';
import { ChartBarIcon, ClockIcon, ExclamationCircleIcon } from '@heroicons/react/24/outline';
import React, { useState } from 'react';
import { PluginWithStats } from '../types/plugin';

interface PluginCardProps {
  plugin: PluginWithStats;
  onStatusChange: (isActive: boolean) => Promise<void>;
  onConfigChange: (config: Record<string, any>) => Promise<void>;
}

export const PluginCard: React.FC<PluginCardProps> = ({
  plugin,
  onStatusChange,
  onConfigChange,
}) => {
  const [isLoading, setIsLoading] = useState(false);

  const handleStatusChange = async (checked: boolean) => {
    setIsLoading(true);
    try {
      await onStatusChange(checked);
    } catch (error) {
      console.error('Failed to update plugin status:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6 space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-gray-900">{plugin.name}</h3>
          <p className="text-sm text-gray-500">{plugin.description}</p>
        </div>
        <Switch
          checked={plugin.isActive}
          onChange={handleStatusChange}
          disabled={isLoading}
          className={`${plugin.isActive ? 'bg-indigo-600' : 'bg-gray-200'
            } relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2`}
        >
          <span
            className={`${plugin.isActive ? 'translate-x-6' : 'translate-x-1'
              } inline-block h-4 w-4 transform rounded-full bg-white transition-transform`}
          />
        </Switch>
      </div>

      <div className="grid grid-cols-3 gap-4 pt-4 border-t border-gray-100">
        <div className="flex items-center space-x-2">
          <ChartBarIcon className="h-5 w-5 text-gray-400" />
          <div>
            <p className="text-sm font-medium text-gray-900">
              {plugin.stats.eventsProcessed}
            </p>
            <p className="text-xs text-gray-500">Events Processed</p>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <ClockIcon className="h-5 w-5 text-gray-400" />
          <div>
            <p className="text-sm font-medium text-gray-900">
              {plugin.stats.lastProcessed
                ? new Date(plugin.stats.lastProcessed).toLocaleTimeString()
                : 'Never'}
            </p>
            <p className="text-xs text-gray-500">Last Processed</p>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <ExclamationCircleIcon className="h-5 w-5 text-gray-400" />
          <div>
            <p className="text-sm font-medium text-gray-900">
              {plugin.stats.errorCount}
            </p>
            <p className="text-xs text-gray-500">Errors</p>
          </div>
        </div>
      </div>

      {Object.keys(plugin.config).length > 0 && (
        <div className="pt-4 border-t border-gray-100">
          <h4 className="text-sm font-medium text-gray-900 mb-2">Configuration</h4>
          <div className="space-y-2">
            {Object.entries(plugin.config).map(([key, value]) => (
              <div key={key} className="flex items-center space-x-2">
                <label className="text-sm text-gray-700 flex-1">{key}</label>
                <input
                  type={typeof value === 'number' ? 'number' : 'text'}
                  value={value}
                  onChange={(e) =>
                    onConfigChange({
                      ...plugin.config,
                      [key]: e.target.type === 'number' ? Number(e.target.value) : e.target.value,
                    })
                  }
                  className="text-sm rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}; 
