
import { FunctionComponent, ReactNode, useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useBoolean } from '@uifabric/react-hooks'
import {
    Modal,
    DefaultButton,
    PrimaryButton,
    TextField,
} from '@fluentui/react'

import './styles.css'

import * as api from '../../api'

export const useIndexURIModal = (defaultProps: Partial<IndexURIModalProps> = {}): [ JSX.Element, (props: IndexURIModalProps) => any, Function ] => {
    const [ props, setProps ] = useState(defaultProps)
    const [ displayIndexURIModal, {
        setTrue: openIndexURIModal,
        setFalse: closeIndexURIModal,
    } ] = useBoolean(false)

    const Component = (
        <IndexURIModal
            isOpen={displayIndexURIModal}
            onClose={closeIndexURIModal}
            {...defaultProps}
            {...props}
        />
    )

    const handleOpenIndexURIModal = (props: IndexURIModalProps) => {
        setProps(props)
        openIndexURIModal()
    }
    
    const handleCloseIndexURIModal = () => {
        closeIndexURIModal()
        setProps({})
    }

    return [ Component, handleOpenIndexURIModal, handleCloseIndexURIModal ]
}

export interface IndexURIModalProps {
    isOpen?: boolean
    onSave?(source: api.protocol.Source)
    onClose?()
}

export const IndexURIModal = ({ isOpen, onSave, onClose }: IndexURIModalProps) => {
    const { t } = useTranslation(['risp', 'modal.index_uri', 'verbs'])

    const [ isIndexing, setIsIndexing ] = useState(false)
    const [ uri, setURI ] = useState('')

    const handleIndex = useCallback(async() => {
        setIsIndexing(true)

        if (!/^(file:\/\/)|(http(s?):\/\/)/i.test(uri)) {
            console.error('Invalid URI')
            alert('Invalid URI')
            return
        }

        try {
            const response = await api.IndexURI(uri)

            if (response?.error?.code) {
                throw response
            }

            if (typeof onSave === 'function') {
                onSave(response.source)
            }
        } catch (err) {
            console.error(err)
            alert(err)
        } finally {
            setIsIndexing(false)

            if (typeof onClose === 'function') {
                onClose()
            }
        }
    }, [ uri ])

    return (
        <Modal
            containerClassName='index-uri-modal-container'
            isOpen={isOpen}
        >
            <div style={{ marginBottom: '24px' }}>
                <h3>{t('modal.index_uri.title')}</h3>
            </div>
            <TextField
                label={t('uri')}
                placeholder={t('modal.index_uri.PlaceholderURI')}
                value={uri}
                onChange={(event: any) =>
                    setURI(event?.target?.value)}
            />
            <div style={{ display: 'flex', flexDirection: 'row', justifyContent: 'right', marginTop: '24px' }}>
                <DefaultButton onClick={onClose}>
                    {t('Close')}
                </DefaultButton>
                <PrimaryButton onClick={handleIndex}>
                    {t('verbs.Index')}
                </PrimaryButton>
            </div>
        </Modal>
    )
}
