
import { useCallback, useEffect, useState, useRef } from 'react'
import { RouteProps, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
    PrimaryButton,
    CommandBar,
    DetailsList,
    DetailsListLayoutMode,
    Selection,
    FontIcon,
} from '@fluentui/react'

import './styles.css'

import { RispAdapterType } from '../../../types'
import * as api from '../../api'
import { useIndexURIModal, IndexURIModal } from '../../modals/indexURI'

export interface SourcesScreenProps extends RouteProps {}
export const SourcesScreen = ({}: SourcesScreenProps) => {
    const navigate = useNavigate()
    const { t } = useTranslation(['risp', 'screen.sources'])

    const [ response, setResponse ] = useState<api.protocol.GetSourcesResponse>(null)

    const [ isCompact, setIsCompact ] = useState(false)
    const [ selectionDetails, setSelectionDetails ] = useState('No sources selected')
    const [ selection ] = useState(new Selection({ onSelectionChanged: () => setSelectionDetails(getSelectionDetails()) }))

    const getSelectionDetails = useCallback((): string => {
        const selectionCount = selection.getSelectedCount()

        let selectionDetails = t('SelectionCount', {
            count: selectionCount,
            name: t('sources', { count: selectionCount }),
        })

        if (selectionCount === 1) {
            selectionDetails = `${selectionDetails}: ${(selection.getSelection()[0] as api.protocol.Source).canonical_uri}`
        }

        return selectionDetails
    }, [ selection ])

    const loadSources = async() => {
        try {
            const response = await api.GetSources()

            setResponse(response)

            console.log('loaded sources:', response)
        } catch (err) {
            console.error(err)
        }
    }

    useEffect(() => {
        loadSources()
    }, [])

    const [ indexURIModal, openIndexURIModal ] = useIndexURIModal({
        onSave: loadSources,
    })

    return (
        <div className='sources-screen'>
            <CommandBar
                items={[{
                    key: 'index_uri',
                    text: t('IndexURI'),
                    iconProps: { iconName: 'Edit' },
                    onClick: () =>
                        openIndexURIModal({}),
                }, {
                    key: 'import_sources',
                    text: t('screen.sources:ImportSources'),
                    iconProps: { iconName: 'Export' },
                }, {
                    key: 'export_sources',
                    text: t('screen.sources:ExportSources'),
                    iconProps: { iconName: 'Import' },
                    onClick: () => {
                        // if (inputElement?.current) {
                        //     inputElement.current.click()
                        // }
                    },
                }/*{
                    key: 'new',
                    text: 'New',
                    // cacheKey: 'foo', // changing this key will invalidate this item's cache
                    iconProps: { iconName: 'Add' },
                    subMenuProps: {
                        items: [{
                            key: 'index_uri',
                            text: 'Index URI',
                            iconProps: { iconName: 'Edit' },
                            onClick: () =>
                                openIndexURIModal({}),
                        }, {
                            key: 'ImportSources',
                            text: 'Import sources',
                            iconProps: { iconName: 'Import' },
                        }],
                    },
                }*/]}
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
                items={response?.sources || []}
                columns={[{
                    key: 'adapter',
                    name: t('Adapter'),
                    onRender: (source: api.protocol.Source) => {
                        const renderAdapterIcon = (iconName: string, label: string) => {
                            return [
                                <FontIcon
                                    key='icon'
                                    iconName={iconName}
                                    style={{
                                        marginRight: '.5rem',
                                        verticalAlign: 'middle',
                                        height: '16px',
                                        width: '16px',
                                    }}
                                />,
                                label,
                            ]
                        }

                        switch (source.adapter_type) {
                        case RispAdapterType.FS:
                        case undefined:
                            return renderAdapterIcon('HardDriveGroup', 'FS')
                        case RispAdapterType.WEB:
                            return renderAdapterIcon('Globe2', 'Web')
                        }

                        return '-'
                    },
                    minWidth: 32,
                    maxWidth: 64,
                    isResizable: true,
                }, {
                    key: 'canonical_uri',
                    name: t('CanonicalURI'),
                    fieldName: 'canonical_uri',
                    minWidth: 100,
                    maxWidth: 200,
                    isResizable: true,
                }, {
                    key: 'urn',
                    name: t('urn'),
                    fieldName: 'urn',
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
            {indexURIModal}
        </div>
    )
}
