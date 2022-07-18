
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

export const useCreateContextModal = (defaultProps: Partial<CreateContextModalProps> = {}): [ JSX.Element, (props: CreateContextModalProps) => any, Function ] => {
    const [ props, setProps ] = useState(defaultProps)
    const [ displayCreateContextModal, {
        setTrue: openCreateContextModal,
        setFalse: closeCreateContextModal,
    } ] = useBoolean(false)

    const Component = (
        <CreateContextModal
            isOpen={displayCreateContextModal}
            onClose={closeCreateContextModal}
            {...defaultProps}
            {...props}
        />
    )

    const handleOpenCreateContextModal = (props: CreateContextModalProps) => {
        setProps(props)
        openCreateContextModal()
    }
    
    const handleCloseCreateContextModal = () => {
        closeCreateContextModal()
        setProps({})
    }

    return [ Component, handleOpenCreateContextModal, handleCloseCreateContextModal ]
}

export interface CreateContextModalProps {
    isOpen?: boolean
    onSave?(context: api.protocol.Context)
    onClose?()
}

export const CreateContextModal = ({ isOpen, onSave, onClose }: CreateContextModalProps) => {
    const { t } = useTranslation(['risp', 'modal.create_context'])

    const [ isSaving, setIsSaving ] = useState(false)
    const [ name, setName ] = useState('')

    const handleSave = useCallback(async() => {
        setIsSaving(true)

        if (/^\s*$/.test(name)) {
            console.error('Invalid name')
            alert('Invalid name')
            return
        }

        try {
            const response = await api.CreateContext(name)

            if (response?.error?.code) {
                throw response
            }

            if (typeof onSave === 'function') {
                onSave(response.context)
            }
        } catch (err) {
            console.error(err)
            alert(err)
        } finally {
            setIsSaving(false)

            if (typeof onClose === 'function') {
                onClose()
            }
        }
    }, [ name ])

    return (
        <Modal
            containerClassName='create-context-modal-container'
            isOpen={isOpen}
        >
            <div style={{ marginBottom: '24px' }}>
                <h3>{t('modal.create_context.title')}</h3>
            </div>
            <TextField
                label='Name'
                placeholder={t('modal.create_context.PlaceholderName')}
                value={name}
                onChange={(event: any) =>
                    setName(event?.target?.value)}
            />
            <div style={{ display: 'flex', flexDirection: 'row', justifyContent: 'right', marginTop: '24px' }}>
                <DefaultButton onClick={onClose}>
                    {t('Close')}
                </DefaultButton>
                <PrimaryButton onClick={handleSave}>
                    {t('Save')}
                </PrimaryButton>
            </div>
        </Modal>
    )
}
