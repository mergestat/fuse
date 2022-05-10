import { Filter, Input } from '@mergestat/blocks'
import { SearchIcon } from '@mergestat/icons'
import React from 'react'

export const ReposTableFilterHeader: React.FC = () => {
  return (
    <div className="flex h-14 bg-white items-center justify-between px-8">
      <div className="flex gap-2">
        <Filter>Tags</Filter>
        <Filter>Filter label</Filter>
        <Filter>Filter label</Filter>
      </div>
      <label className="">
        <SearchIcon className="t-icon absolute left-2 text-gray-400" />
        <Input placeholder="Search..." className="pl-8" />
      </label>
    </div>
  )
}