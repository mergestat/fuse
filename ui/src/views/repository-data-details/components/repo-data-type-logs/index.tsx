import React from 'react'
import { Button, Panel } from '@mergestat/blocks'
import { LogsTable } from './logs-table'

export const RepoDataLogs = ({ data }: { data: any}) => {
  return (
    <div className="h-full">
      <Panel className="shadow-sm mb-8">
        <Panel.Body>
          <h4 className="t-h4 mb-2">
            {data.title}
          </h4>
          <p className="text-semantic-mutedText">
            {data.brief}
          </p>

          <Button skin="borderless" label="Learn more" />
        </Panel.Body>
      </Panel>

      <div className='border-md shadow-sm'>
        <LogsTable />
      </div>
    </div>
  )
}
