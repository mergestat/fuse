import React, { useEffect, useState } from 'react'
import cx from 'classnames'
import { RepoSyncDataType } from 'src/@types'
import { getTimeAgoFromNow } from 'src/utils'

export type TimeAgoFieldProps = {
  date: string,
  syncData?: RepoSyncDataType,
  extraStyles?: string
}

export const TimeAgoField: React.FC<TimeAgoFieldProps> = ({ date, syncData, extraStyles }) => {
  const [dateRun, setDateRun] = useState<string>('N/A')

  const recalculateTimeAgo = () => {
    const timeAgo = getTimeAgoFromNow(new Date(date))
    setDateRun(timeAgo)
  }

  useEffect(() => {
    let intervalId: NodeJS.Timer
    if (date !== '') {
      recalculateTimeAgo()
      intervalId = setInterval(recalculateTimeAgo, 30000)
    }
    return () => clearInterval(intervalId);
  }, [])

  return (
    <div
      className={cx("text-semantic-mutedText", extraStyles, {
        'bg-gray-50': syncData?.status.syncState === 'disabled'
      })}
    >
      {dateRun}
    </div>
  )
}
