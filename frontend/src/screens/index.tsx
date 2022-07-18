
import { useCallback, useState } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'

import './styles.css'

import { TopNavBar, topNavBarContext } from '../components/topNavBar'
import { SearchScreen } from './search'
import { SearchResultsScreen } from './search/results'
import { ResourcesScreen } from './resources'
import { SourcesScreen } from './sources'
import { ContextsScreen } from './contexts'

export const Screens = () => {
    const [ isMenuOpen, setIsMenuOpen ] = useState(true)
    const toggleMenu = useCallback(() => setIsMenuOpen(!isMenuOpen), [ isMenuOpen ])

    return (
        <div className='screens'>
            <topNavBarContext.Provider value={{ isMenuOpen, toggleMenu }}>
                <TopNavBar/>
                <Routes>
                    <Route path='/' element={<Navigate to='/search' replace />} />
                    <Route path='/search' element={<SearchScreen/>} />
                    <Route path='/search/results' element={<SearchResultsScreen/>} />
                    <Route path='/resources' element={<ResourcesScreen/>} />
                    <Route path='/sources' element={<SourcesScreen/>} />
                    <Route path='/contexts' element={<ContextsScreen/>} />
                </Routes>
            </topNavBarContext.Provider>
        </div>
    )
}
