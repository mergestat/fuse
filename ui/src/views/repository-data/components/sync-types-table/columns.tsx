import React from 'react'
import cx from 'classnames'
import { RepoSyncStateT } from 'src/@types'
import type { RepoSyncDataType, SyncStatusDataT } from 'src/@types'
import { RepoSyncIcon } from 'src/components/RepoSyncIcon'
import {
  RepositoryData,
  RepositoryDataProps,
  RepositorySyncNow,
  RepositorySyncStatus,
  RepositoryTableRowOptions,
} from './components'
import { TimeAgoField } from 'src/components/Fields/time-ago-field'

export const columns: Array<Record<string, any>> = [
  {
    dataIndex: 'status',
    key: 'syncStateIcon',
    className: "w-12 h-20 p-0",
    render: ({ syncState }: { syncState: RepoSyncStateT }) => (
      <div className={cx('h-full px-6 flex', { 'bg-gray-50': syncState === 'disabled' })}>
        <RepoSyncIcon type={syncState} className="my-auto" />
      </div>
    )
  },
  {
    title: 'Data',
    dataIndex: 'data',
    headerClassName: "pl-0",
    key: 'data',
    className: "h-20 p-0",
    render: (data: RepositoryDataProps, d: RepoSyncDataType) => (
      <RepositoryData
        title={data.title}
        brief={data.brief}
        disabled={d.status.syncState === 'disabled'}
      />
    )
  },
  {
    title: 'Lastest Run',
    headerClassName: "pl-0",
    className: 'text-gray-500 h-20 p-0',
    dataIndex: 'latestRun',
    key: 'latestRun',
    render: (latestRun: string, d: RepoSyncDataType) =>
      <TimeAgoField date={latestRun} syncData={d} extraStyles={'h-full leading-20'} />
  },
  {
    title: 'Status',
    headerClassName: "pl-0",
    className: 'text-gray-500 h-20 p-0',
    dataIndex: 'status',
    key: 'status',
    render: (status: { data?: SyncStatusDataT[], syncState: RepoSyncStateT }) =>
      <RepositorySyncStatus
        data={status.data}
        disabled={status.syncState === 'disabled'}
      />
  },
  {
    dataIndex: 'status',
    className: 'h-20 p-0',
    key: 'syncNow',
    render: ({ syncState }: { syncState: RepoSyncStateT }) =>
      <RepositorySyncNow syncStatus={syncState} />
  },
  {
    className: 'w-4 p-0 h-20',
    dataIndex: 'status',
    key: 'options',
    render: ({ syncState }: { syncState: RepoSyncStateT }) => (
      <div className={cx('h-full flex', { 'bg-gray-50': syncState === 'disabled' })}>
        <div className='my-auto mx-6'>
          <RepositoryTableRowOptions state={syncState} />
        </div>
      </div>
    ),
  },
]
