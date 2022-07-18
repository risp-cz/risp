
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

import { RispResourceType } from '../../../types'
import * as api from '../../api'
import { useIndexURIModal, IndexURIModal } from '../../modals/indexURI'

const knownFileTypes = [
    'accdb', 'audio', 'code', 'csv', 'docx', 'dotx', 'mpp', 'mpt', 'model', 'one', 'onetoc', 'potx', 'ppsx', 'pdf',
    'photo', 'pptx', 'presentation', 'potx', 'pub', 'rtf', 'spreadsheet', 'txt', 'vector', 'vsdx', 'vssx', 'vstx',
    'xlsx', 'xltx', 'xsn',
]
const getFileTypeIconURL = (type: string) => {
    return `https://static2.sharepointonline.com/files/fabric/assets/item-types/16/${type}.svg`
}

export interface ResourcesScreenProps extends RouteProps {}
export const ResourcesScreen = ({}: ResourcesScreenProps) => {
    const navigate = useNavigate()
    const { t } = useTranslation(['risp', 'screen.resources'])

    const [ response, setResponse ] = useState<api.protocol.GetResourcesResponse>(null)

    const [ isCompact, setIsCompact ] = useState(false)
    const [ selectionDetails, setSelectionDetails ] = useState('No resources selected')
    const [ selection ] = useState(new Selection({ onSelectionChanged: () => setSelectionDetails(getSelectionDetails()) }))

    const getSelectionDetails = useCallback((): string => {
        const selectionCount = selection.getSelectedCount()

        let selectionDetails = t('SelectionCount', {
            count: selectionCount,
            name: t('resources', { count: selectionCount }),
        })

        if (selectionCount === 1) {
            selectionDetails = `${selectionDetails}: ${(selection.getSelection()[0] as api.protocol.Resource).canonical_uri}`
        }

        return selectionDetails
    }, [ selection ])

    const loadResources = async() => {
        try {
            const response = await api.GetResources()

            setResponse(response)

            console.log('loaded resources:', response)
        } catch (err) {
            console.error(err)
        }
    }

    useEffect(() => {
        loadResources()
    }, [])

    const [ indexURIModal, openIndexURIModal ] = useIndexURIModal({
        onSave: loadResources,
    })

    return (
        <div className='resources-screen'>
            <CommandBar
                items={[{
                    key: 'index_uri',
                    text: t('IndexURI'),
                    iconProps: { iconName: 'Edit' },
                    onClick: () =>
                        openIndexURIModal({}),
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
                items={response?.resources || []}
                columns={[{
                    key: 'type',
                    name: t('Type'),
                    onRender: (resource: api.protocol.Resource) => {
                        switch (resource.type) {
                        case RispResourceType.FS_FILE:
                        case undefined:
                            return (
                                <FontIcon
                                    iconName='Page'
                                    style={{
                                        height: '16px',
                                        width: '16px',
                                    }}
                                />
                            )
                        case RispResourceType.WEB_PAGE:
                            return (
                                <FontIcon
                                    iconName='Globe'
                                    style={{
                                        height: '16px',
                                        width: '16px',
                                    }}
                                />
                            )
                        }

                        return (
                            <img
                                src={`https://static2.sharepointonline.com/files/fabric/assets/item-types/16/${'model'}.svg`}
                                style={{
                                    verticalAlign: 'middle',
                                    maxHeight: '16px',
                                    maxWidth: '16px',
                                }}
                                alt=''
                            />
                        )

                        switch (resource.type) {
                        case RispResourceType.FS_FILE:
                        case undefined:
                            return 'File'
                        case RispResourceType.WEB_PAGE:
                            return 'Web page'
                        }

                        return '-'
                    },
                    minWidth: 32,
                    maxWidth: 64,
                    isResizable: true,
                }, /*{
                    key: 'source_urn',
                    name: 'Source URN',
                    fieldName: 'source_urn',
                    minWidth: 100,
                    maxWidth: 200,
                    isResizable: true,
                },*/ {
                    key: 'source_canonical_uri',
                    name: t('screen.resources:SourceCanonicalURI'),
                    fieldName: 'source_canonical_uri',
                    minWidth: 100,
                    maxWidth: 200,
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
