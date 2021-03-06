import React from 'react'
import type { RepoDataPropsT, RepoDataStatusT, TagType } from 'src/@types'
import { TimeAgoField } from 'src/components/Fields/time-ago-field'
import { RepositoryAdditionalActionsDropDown } from '../../drop-downs'
import {
  RepositoryName,
  RepositoryStatus,
  RepositoryTagList,
} from './repositories-table-columns'

export const columns: Array<Record<string, any>> = [
  {
    title: 'Repository',
    dataIndex: 'name',
    key: 'name',
    onSortChange: (e: 'asc' | 'desc' | undefined) => {
      console.log(e)
    },
    render: (name: string, data: RepoDataPropsT) => (
      <RepositoryName
        id={data.id}
        name={name}
        type={data.type}
        createdAt={data.createdAt}
        automaticImport={data.automaticImport}
      />
    ),
  },
  {
    title: 'Tags',
    className: 'text-gray-500',
    dataIndex: 'tags',
    key: 'tags',
    render: (tags: TagType[]) => <RepositoryTagList tags={tags} />,
  },
  {
    title: 'Last sync',
    dataIndex: 'lastSync',
    className: 'h-20',
    key: 'last',
    onSortChange: (e: 'asc' | 'desc' | undefined) => {
      console.log(e)
    },
    render: (lastSync: string) => <TimeAgoField date={lastSync} extraStyles={'px-6'} />
  },
  {
    dataIndex: 'status',
    key: 'status',
    className: 'h-20',
    render: (status: Array<RepoDataStatusT>) => (
      <RepositoryStatus status={status} />
    ),
  },
  {
    className: 'px-6 w-4',
    key: 'option',
    render: () => <RepositoryAdditionalActionsDropDown />,
  },
]
