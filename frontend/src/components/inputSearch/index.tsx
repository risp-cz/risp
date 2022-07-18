
import { useEffect, useState } from 'react'
import { RouteProps } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

import './styles.css'

export interface InputSearchProps extends RouteProps {
    defaultValue?: string
    onSearch?(searchExpression: string): void
}

export const InputSearch = ({ defaultValue, onSearch }: InputSearchProps) => {
    const { t } = useTranslation(['risp', 'verbs'])

    const [ searchExpression, setSearchExpression ] = useState(defaultValue || '')

    useEffect(() => {
        setSearchExpression(defaultValue)
    }, [defaultValue])

    const handleSearch = () => {
        if (typeof onSearch === 'function') {
            onSearch(searchExpression)
        }
    }

    const handleKeyDown = (event) => {
        switch (String(event.key).toLowerCase()) {
        case 'enter':
            handleSearch()
            return
        }
    }

    return (
        <div className='input-search-container'>
            <input
                className='input-search'
                placeholder={`${t('verbs:SearchFor')}...`}
                value={searchExpression}
                onChange={event => setSearchExpression(event.target.value)}
                onKeyDown={handleKeyDown}
            />
        </div>
    )
}
