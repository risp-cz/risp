
import { createContext, useEffect, useState, useContext } from 'react'
import { RouteProps, useNavigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
    Layer,
    Pivot,
    PivotItem,
    FontIcon,
} from '@fluentui/react'

import './styles.css'

export const topNavBarContext = createContext({
    isMenuOpen: true,
    toggleMenu: () => {},
})

export interface MenuItem {
    title: string
    path: string
}

export interface TopNavBarProps extends RouteProps {}
export const TopNavBar = ({}: TopNavBarProps) => {
    const { isMenuOpen, toggleMenu } = useContext(topNavBarContext)

    const { t } = useTranslation(['risp', 'verbs'])

    const navigate = useNavigate()
    const location = useLocation()

    const menuItems: MenuItem[] = [{
        title: t('verbs:Search'),
        path: '/search',
    }, {
        title: t('Resource', { count: 0 }),
        path: '/resources',
    }, {
        title: t('Source', { count: 0 }),
        path: '/sources',
    }, {
        title: t('Context', { count: 0 }),
        path: '/contexts',
    }]

    const handleToggleMenu = () => {
        if (typeof toggleMenu === 'function') {
            toggleMenu()
        }
    }

    const renderMenuItem = ({ title , path}: MenuItem, index: number) => {
        return (
            <PivotItem
                key={`${index}${path}`}
                headerText={title}
                itemKey={path}
            />
        )
    }

    return (
        <Layer>
            <div className={`top-nav-bar-container ${isMenuOpen ? 'opened' : 'closed'}`}>
                <div className='top-nav-bar-menu'>
                    <div className='top-nav-bar-menu-items'>
                        <Pivot
                            overflowBehavior='menu'
                            selectedKey={menuItems.reduce<string>((selected, item) => {
                                return selected || (location.pathname.startsWith(item.path) ? item.path : null)
                            }, null)}
                            onLinkClick={({ props }) =>
                                navigate(`${props.itemKey}`)}
                            headersOnly
                        >
                            {menuItems.map(renderMenuItem)}
                        </Pivot>
                    </div>
                </div>
                <div
                    className='top-nav-bar-toggle'
                    onClick={handleToggleMenu}
                >
                    <FontIcon iconName='GlobalNavButton' />
                </div>
            </div>
        </Layer>
    )
}
