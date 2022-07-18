
import { useState } from 'react'
import { RouteProps, useNavigate } from 'react-router-dom'

import './styles.css'

import { InputSearch } from '../../components/inputSearch'

export interface SearchScreenProps extends RouteProps {}
export const SearchScreen = ({}: SearchScreenProps) => {
    const navigate = useNavigate()

    return (
        <div className='search-screen'>
            <div className='search-screen-title font-handwriting'>
                Risp search
            </div>
            <div className='search-screen-input'>
                <InputSearch
                    defaultValue=''
                    onSearch={searchExpression =>
                        navigate(`/search/results?query=${encodeURIComponent(searchExpression)}`)}
                />
            </div>
        </div>
    )
}
