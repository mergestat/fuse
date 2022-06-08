import { Tabs } from '@mergestat/blocks'
import React from 'react'
import type { RepoDataPropsT } from 'src/@types'
import {
  PageHeader,
  RepoSettings,
  SycnTypesTable,
} from './components'

const RepoDataView = ({ data }: { data: RepoDataPropsT }) => {
  return (
    <main className="w-full flex flex-col h-full overflow-hidden bg-gray-50">
      <PageHeader data={data} />
      <Tabs.Group>
        <Tabs.List className="bg-white w-full justify-between px-8 items-center border-b border-gray-200">
          <Tabs.Item className="ring-transparent focus_ring-transparent">
            Sync Types
          </Tabs.Item>
          <Tabs.Item>Repo Settings</Tabs.Item>
        </Tabs.List>
        <Tabs.Panels className="p-8 flex-1 overflow-auto">
          <Tabs.Panel>
            <SycnTypesTable />
          </Tabs.Panel>
          <Tabs.Panel>
            <RepoSettings />
          </Tabs.Panel>
        </Tabs.Panels>
      </Tabs.Group>

      {/*false && <EmptyRepositoryTable />*/}
      {/*true && */}
    </main>
  )
}

export default RepoDataView
