
import { useCallback, useEffect, useState } from 'react'
import { RouteProps, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
    PrimaryButton,
    CommandBar,
    DetailsList,
    DetailsListLayoutMode,
    Selection,
} from '@fluentui/react'

import './styles.css'

import * as api from '../../api'
import { useCreateContextModal } from '../../modals/createContext'

export interface ContextsScreenProps extends RouteProps {}
export const ContextsScreen = ({}: ContextsScreenProps) => {
    const navigate = useNavigate()
    const { t } = useTranslation(['risp', 'screen.contexts'])

    const [ response, setResponse ] = useState<api.protocol.GetContextsResponse>(null)

    const [ isCompact, setIsCompact ] = useState(false)
    const [ selectionDetails, setSelectionDetails ] = useState('No contexts selected')
    const [ selection ] = useState(new Selection({ onSelectionChanged: () => setSelectionDetails(getSelectionDetails()) }))
    const selectionCount = selection.getSelectedCount()

    const getSelectionDetails = useCallback((): string => {
        const selectionCount = selection.getSelectedCount()

        let selectionDetails = t('SelectionCount', {
            count: selectionCount,
            name: t('contexts', { count: selectionCount }),
        })

        if (selectionCount === 1) {
            selectionDetails = `${selectionDetails}: ${(selection.getSelection()[0] as api.protocol.Context).name}`
        }

        return selectionDetails
    }, [ selection ])

    const loadContexts = async() => {
        try {
            setResponse(await api.GetContexts())
        } catch (err) {
            console.error(err)
        }
    }

    useEffect(() => {
        loadContexts()
    }, [])

    const [ createContextModal, openCreateContextModal ] = useCreateContextModal({
        onSave: loadContexts,
    })

    const handleExportSelected = async() => {
        try {
            const contextIds = selection.getSelection().map(({ id }) => id)
            const response = await api.ExportContexts(contextIds)
            console.log('response:', response)
        } catch (err) {
            console.error(err)
        }
    }

    return (
        <div className='contexts-screen'>
            <CommandBar
                items={[{
                    key: 'new',
                    text: t('New'),
                    iconProps: { iconName: 'CircleAddition' },
                    onClick: () =>
                        openCreateContextModal({}),
                }, {
                    key: 'export',
                    text: `${t('Export')}${selectionCount < 1 ? '' : ` ${selectionCount} ${t('contexts', { count: selectionCount })}`}`,
                    iconProps: { iconName: 'Import' },
                    disabled: selectionCount < 1,
                    onClick: () => {
                        handleExportSelected()
                        return true
                    },
                }]}
                overflowItems={[{
                    key: 'delete',
                    text: t('Delete'),
                    onClick: () => {},
                    iconProps: { iconName: 'Delete' },
                }]}
                overflowButtonProps={{ ariaLabel: 'More commands' }}
                farItems={[{
                    key: 'compact',
                    text: t('CompactView'),
                    ariaLabel: t('CompactView'),
                    iconOnly: true,
                    iconProps: { iconName: isCompact ? 'TransitionPop' : 'TransitionPush' },
                    onClick: () =>
                        setIsCompact(!isCompact),
                }]}
                // ariaLabel="Use left and right arrow keys to navigate between commands"
            />
            <div style={{ textAlign: 'right', marginRight: '16px' }}>
                {selectionDetails}
            </div>
            <DetailsList
                items={response?.contexts || []}
                columns={[{
                    key: 'id',
                    name: t('id'),
                    fieldName: 'id',
                    minWidth: 128,
                    maxWidth: 256,
                    isResizable: true,
                }, {
                    key: 'is_default',
                    name: t('Default'),
                    fieldName: 'is_default',
                    minWidth: 32,
                    maxWidth: 64,
                    isResizable: true,
                }, {
                    key: 'name',
                    name: t('Name'),
                    fieldName: 'name',
                    minWidth: 100,
                    maxWidth: 200,
                    isResizable: true,
                }]}
                setKey='set'
                layoutMode={DetailsListLayoutMode.justified}
                onItemInvoked={(item) => console.log('item invoked:', item)}
                selection={selection}
                compact={isCompact}
                selectionPreservedOnEmptyClick
                // ariaLabelForSelectionColumn="Toggle selection"
                // ariaLabelForSelectAllCheckbox="Toggle selection for all items"
                // checkButtonAriaLabel="Row checkbox"
            />
            {createContextModal}
        </div>
    )
}
