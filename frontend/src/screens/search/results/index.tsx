
import { useEffect, useState } from 'react'
import { RouteProps, useNavigate, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

import './styles.css'

import { RispResourceType } from '../../../../types'
import * as api from '../../../api'
import { InputSearch } from '../../../components/inputSearch'

const knownFileTypes = [
    'accdb', 'audio', 'code', 'csv', 'docx', 'dotx', 'mpp', 'mpt', 'model', 'one', 'onetoc', 'potx', 'ppsx', 'pdf',
    'photo', 'pptx', 'presentation', 'potx', 'pub', 'rtf', 'spreadsheet', 'txt', 'vector', 'vsdx', 'vssx', 'vstx',
    'xlsx', 'xltx', 'xsn',
]
const getFileTypeIconURL = (type: string) => {
    return `https://static2.sharepointonline.com/files/fabric/assets/item-types/16/${type}.svg`
}

export interface SearchResultsScreenProps extends RouteProps {}
export const SearchResultsScreen = ({}: SearchResultsScreenProps) => {
    const { t } = useTranslation(['risp', 'screen.search'])
    const navigate = useNavigate()
    const [ searchParams ] = useSearchParams()
    const query = searchParams.get('query')

    const [ result, setResult ] = useState<api.protocol.QueryResponse>(null)

    const handleSearch = async(searchExpression: string) => {
        try {
            const result = await api.Query(searchExpression)
            console.log('query result:', result)

            if (result?.error?.code) {
                alert(result.error)
                return
            }

            setResult(result)
        } catch (err) {
            console.log(err)
        }
    }

    useEffect(() => {
        handleSearch(query)
    }, [query])

    const renderResultFSFile = ({ score, resource, highlights }: api.protocol.QueryHit, index: number) => {
        let preview = null
        let imageUrl = null
        let iconName = null

        const resourceUri = `${resource.source_canonical_uri.replace(/\s/g, '%20')}/${resource.canonical_uri.replace(/\s/g, '%20')}`
        const data: {
            path: string
            filename: string
            filetype: string
            isDot: boolean
        } = JSON.parse(resource.data_json)

        if (['png', 'jpg', 'jpeg'].indexOf(data?.filetype) >= 0) {
            data.filetype = 'photo'
        }

        if (knownFileTypes.indexOf(data?.filetype) >= 0) {
            imageUrl = getFileTypeIconURL(data.filetype)
        }

        for (const highlight of highlights || []) {
            if (highlight.key === 'fs-file.contents_text') {
                if (highlight?.values?.length > 0)  {
                    preview = highlight.values[0]
                }

                continue
            }

            if (!preview && highlight.key === 'fs-file.contents_html') {
                if (highlight?.values?.length > 0)  {
                    preview = highlight.values[0]
                }

                continue
            }
        }

        return (
            <div
                key={`${index}${resource.urn}`}
                className='search-result-container'
            >
                <div className='search-result-url'>
                    {data?.path || resourceUri}
                </div>
                {imageUrl && (
                    <div className='search-result-image'>
                        <img
                            src={imageUrl}
                            style={{ verticalAlign: 'middle' }}
                            alt={data?.filename}
                        />
                    </div>
                )}
                {(!imageUrl && iconName) && (
                    <div className='search-result-image'>
                    </div>
                )}
                <a
                    className='search-result-title'
                    href={resourceUri}
                    onClick={async(event) => {
                        event.preventDefault()
                        event.stopPropagation()

                        try {
                            const error = await api.OpenURI(resourceUri)

                            if (error) {
                                throw error
                            }
                        } catch (err) {
                            console.error(err)
                        }
                    }}
                >
                    {resource.canonical_uri}
                </a>
                <div
                    className='search-result-preview'
                    dangerouslySetInnerHTML={{ __html: preview || '&nbsp;' }}
                />
            </div>
        )
    }

    const renderResultWebPage = ({ score, resource, highlights }: api.protocol.QueryHit, index: number) => {
        let preview = null
        let titleHighlight = null

        for (const highlight of highlights || []) {
            if (highlight.key === 'web-page.body') {
                if (highlight?.values?.length > 0)  {
                    preview = highlight.values[0]
                }

                continue
            }

            if (highlight.key === 'web-page.title') {
                if (highlight?.values?.length > 0)  {
                    titleHighlight = highlight.values[0]
                }

                continue
            }
        }

        return (
            <div
                key={`${index}${resource.urn}`}
                className='search-result-container'
            >
                <div className='search-result-url'>
                    {`${resource.source_canonical_uri}${resource.canonical_uri}`}
                </div>
                <a
                    className='search-result-title'
                    href={`${resource.source_canonical_uri}${resource.canonical_uri}`}
                    dangerouslySetInnerHTML={{ __html: titleHighlight }}
                />
                {preview && (
                    <div
                        className='search-result-preview'
                        dangerouslySetInnerHTML={{ __html: preview }}
                    />
                )}
            </div>
        )
    }

    const renderResult = (hit: api.protocol.QueryHit, index: number) => {
        switch (hit?.resource?.type) {
        case RispResourceType.FS_FILE:
        case undefined:
            return renderResultFSFile(hit, index)
        case RispResourceType.WEB_PAGE:
            return renderResultWebPage(hit, index)
        }

        return '-'
    }

    return (
        <div className='search-results-screen'>
            <div className='search-results-query'>
                <InputSearch
                    defaultValue={query}
                    onSearch={searchExpression =>
                        navigate(`/search/results?query=${encodeURIComponent(searchExpression)}`)}
                />
            </div>
            <div className='search-results-info'>
                <div className='search-results-info-query'>{t('screen.search:ShowingResultsFor', { query })}</div>
                <div className='search-results-info-count'>{t('screen.search:total', { total: result?.edges_total || 0 })}</div>
            </div>
            {result?.edges_total > 0 ? (
                (result.edges || []).map(renderResult)
            ) : (
                <div className='search-results-empty'>{t('screen.search:NoMatch')}</div>
            )}
        </div>
    )
}
